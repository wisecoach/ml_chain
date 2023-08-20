package node

import (
	"encoding/gob"
	"encoding/json"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/discovery"
	"github.com/wisecoach/ml_chain/comm/notary"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"net/rpc"
	"reflect"
	"sync"
	"sync/atomic"
)

type PeerFilter func(peer *proto.RemotePeer) bool

type Node struct {
	notary.NotaryManager
	self   *proto.RemotePeer
	mcs    crypto.MessageCryptoService
	disc   discovery.Discovery
	comm   comm.Comm
	config *Config

	messageListeners map[reflect.Type][]comm.MessageListener

	logger     *zap.Logger
	lock       sync.RWMutex
	toDieChan  chan struct{}
	stopFlag   int32
	stopSignal sync.WaitGroup // wait for stop acceptMessage
}

func New(config *Config, server *rpc.Server, mcs crypto.MessageCryptoService) *Node {
	node := &Node{
		self:             config.Self,
		disc:             discovery.New(config.Self),
		comm:             comm.New(server, config.Self, config.TimeoutRPC),
		config:           config,
		messageListeners: make(map[reflect.Type][]comm.MessageListener),
		logger:           log.GetLogger(config.Self.Endpoint),
		lock:             sync.RWMutex{},
		toDieChan:        make(chan struct{}),
		stopFlag:         int32(0),
		stopSignal:       sync.WaitGroup{},
		mcs:              mcs,
	}

	// init notary manager
	node.NotaryManager = notary.New(node)
	for _, peer := range config.Notaries {
		node.NotaryManager.Discover(peer)
	}
	node.RegisterListener(&proto.NotarySignReqMessage{}, notary.NewNotarySignReqMessageListener(node.mcs, node, config.Self))
	node.RegisterListener(&proto.NotarySignRespMessage{}, notary.NewNotarySignRespMessageListener(node.NotaryManager))

	go node.start()

	return node
}

func (n *Node) Self() *proto.RemotePeer {
	return n.self
}

func (n *Node) Lookup(pk []byte) *proto.RemotePeer {
	return n.disc.Lookup(pk)
}

// Peers
//
//	@Description: get the peers discovered in the network
//	@return *proto.RemotePeer
func (n *Node) Peers() []*proto.RemotePeer {
	return n.disc.GetMembership()
}

// GetDiscovery
//
//	@Description: get the discovery
func (n *Node) GetDiscovery() discovery.Discovery {
	return n.disc
}

func (n *Node) SignMessage(message *proto.Message) (*proto.Envelope[*proto.Message], error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	sign, err := n.mcs.Sign(payload)
	if err != nil {
		return nil, err
	}
	return &proto.Envelope[*proto.Message]{
		Payload:   message,
		Signature: sign,
	}, nil
}

// SendWithFilter
//
//	@Description: send msg to peers filtered by filter
//	@param msg
//	@param filter
//	@return error
func (n *Node) SendWithFilter(msg *proto.Message, filter PeerFilter) {
	peersToSend := filterPeers(n.Peers(), filter)
	envelope, err := n.SignMessage(msg)
	if err != nil {
		n.logger.Error("sign the message failed, err:" + err.Error())
	}
	n.comm.Send(envelope, peersToSend...)
}

func (n *Node) SendToPeers(msg *proto.Message, peers ...*proto.RemotePeer) {
	envelope, err := n.SignMessage(msg)
	if err != nil {
		n.logger.Error("sign the message failed, err:" + err.Error())
	}
	n.comm.Send(envelope, peers...)
}

// RegisterListener
//
//	@Description: register listener to node
func (n *Node) RegisterListener(content any, listener comm.MessageListener) {
	n.lock.Lock()
	defer n.lock.Unlock()

	// register the type to gob
	gob.Register(content)
	contentType := reflect.TypeOf(content)
	listeners := n.messageListeners[contentType]
	if listeners == nil {
		listeners = make([]comm.MessageListener, 0)
	}
	n.messageListeners[contentType] = append(listeners, listener)
}

func (n *Node) start() {
	msgSelector := func(message *proto.ReceivedMessage) bool {
		return true
	}
	messages := n.comm.Accept(msgSelector)

	n.stopSignal.Add(1)
	go n.acceptMessages(messages)
}

func (n *Node) acceptMessages(messages <-chan *proto.ReceivedMessage) {
	defer n.stopSignal.Done()

	for {
		select {
		case <-n.toDieChan:
			return
		case msg := <-messages:
			// n.logger.Info(fmt.Sprintf("get the message from: %s, msg payload type: %T", msg.Sender.Endpoint, msg.Envelope.Payload.Content))
			n.handleMessage(msg)
		}
	}
}

func (n *Node) handleMessage(msg *proto.ReceivedMessage) {
	n.lock.RLock()

	contentType := reflect.TypeOf(msg.Envelope.Payload.Content)
	listeners := n.messageListeners[contentType]

	n.lock.RUnlock()

	for _, listener := range listeners {
		listener.HandleMessage(msg)
	}
}

func filterPeers(peers []*proto.RemotePeer, filter PeerFilter) []*proto.RemotePeer {
	peersToSend := make([]*proto.RemotePeer, 0)
	for _, peer := range peers {
		if filter(peer) {
			peersToSend = append(peersToSend, peer)
		}
	}
	return peersToSend
}

// Stop stops the gossip component
func (n *Node) Stop() {
	if n.toDie() {
		return
	}
	n.logger.Info("node is stopping")
	defer n.logger.Info("node stopped")
	atomic.StoreInt32(&n.stopFlag, int32(1))
	n.stopSignal.Wait()
	close(n.toDieChan)
	n.comm.Stop()

}

func (n *Node) toDie() bool {
	return atomic.LoadInt32(&n.stopFlag) == int32(1)
}
