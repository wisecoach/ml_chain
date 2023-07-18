package node

import (
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/discovery"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/logger"
	"net/rpc"
	"reflect"
	"sync"
	"sync/atomic"
)

type PeerFilter func(peer *comm.RemotePeer) bool

type MessageListener interface {
	HandleMessage(message *comm.ReceivedMessage)
}

type Node struct {
	self   *comm.RemotePeer
	mcs    crypto.MessageCryptoService
	disc   discovery.Discovery
	comm   comm.Comm
	config *Config

	messageListeners map[reflect.Type][]MessageListener

	lock       sync.RWMutex
	toDieChan  chan struct{}
	stopFlag   int32
	stopSignal sync.WaitGroup // wait for stop acceptMessage
}

func New(config *Config, server *rpc.Server) *Node {
	node := &Node{
		self:             config.Self,
		disc:             discovery.New(config.Self),
		comm:             comm.New(server, config.Self, config.TimeoutRPC),
		config:           config,
		messageListeners: make(map[reflect.Type][]MessageListener),
		lock:             sync.RWMutex{},
		toDieChan:        make(chan struct{}),
		stopFlag:         int32(0),
		stopSignal:       sync.WaitGroup{},
	}

	// init the mcs
	bccsp, err := sw.NewBCCSP()
	if err != nil {
		logger.Error("create bccsp failed: " + err.Error())
	}
	node.mcs = crypto.New(bccsp, config.Self, config.KeyImportOpts, config.HashOpts, config.SignerOpts)

	// init the discovery, register the bootstrapPeers
	for _, peer := range config.BootstrapPeers {
		node.disc.Register(peer)
	}

	go node.start()

	return node
}

// Peers
//
//	@Description: get the peers discovered in the network
//	@return *comm.RemotePeer
func (n *Node) Peers() []*comm.RemotePeer {
	return n.disc.GetMembership()
}

// GetDiscovery
//
//	@Description: get the discovery
func (n *Node) GetDiscovery() discovery.Discovery {
	return n.disc
}

func (n *Node) signMessage(message *comm.Message) (*proto.Envelope[*comm.Message], error) {
	panic("impl me")
}

// SendWithFilter
//
//	@Description: send msg to peers filtered by filter
//	@param msg
//	@param filter
//	@return error
func (n *Node) SendWithFilter(msg *comm.Message, filter PeerFilter) {
	peersToSend := filterPeers(n.Peers(), filter)
	envelope, err := n.signMessage(msg)
	if err != nil {
		logger.Error("sign the message failed, err:" + err.Error())
	}
	n.comm.Send(envelope, peersToSend...)
}

func (n *Node) SendToPeers(msg *comm.Message, peers []*comm.RemotePeer) {
	envelope, err := n.signMessage(msg)
	if err != nil {
		logger.Error("sign the message failed, err:" + err.Error())
	}
	n.comm.Send(envelope, peers...)
}

// RegisterListener
//
//	@Description: register listener to node
func (n *Node) RegisterListener(contentType reflect.Type, listener MessageListener) {
	n.lock.Lock()
	defer n.lock.Unlock()

	listeners := n.messageListeners[contentType]
	n.messageListeners[contentType] = append(listeners, listener)
}

func (n *Node) start() {
	msgSelector := func(message *comm.ReceivedMessage) bool {
		return true
	}
	messages := n.comm.Accept(msgSelector)

	n.stopSignal.Add(1)
	go n.acceptMessages(messages)
}

func (n *Node) acceptMessages(messages <-chan *comm.ReceivedMessage) {
	defer n.stopSignal.Done()

	for {
		select {
		case <-n.toDieChan:
			return
		case msg := <-messages:
			n.handleMessage(msg)
		}
	}
}

func (n *Node) handleMessage(msg *comm.ReceivedMessage) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	contentType := reflect.TypeOf(msg.Envelope.Payload.Content)
	listeners := n.messageListeners[contentType]
	for _, listener := range listeners {
		listener.HandleMessage(msg)
	}
}

func filterPeers(peers []*comm.RemotePeer, filter PeerFilter) []*comm.RemotePeer {
	peersToSend := make([]*comm.RemotePeer, 0)
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
	logger.Info("node is stopping")
	defer logger.Info("node stopped")
	atomic.StoreInt32(&n.stopFlag, int32(1))
	n.stopSignal.Wait()
	close(n.toDieChan)
	n.comm.Stop()

}

func (n *Node) toDie() bool {
	return atomic.LoadInt32(&n.stopFlag) == int32(1)
}
