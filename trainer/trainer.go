package trainer

import (
	"github.com/wisecoach/ml_chain/block/chain"
	"github.com/wisecoach/ml_chain/block/consensus"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/message"
	"github.com/wisecoach/ml_chain/trainer/role"
	"github.com/wisecoach/ml_chain/util/logger"
	"net"
	"net/rpc"
	"reflect"
	"time"
)

type Trainer struct {
	self    *comm.RemotePeer
	config  *Config
	chainId string

	roleManager  role.Manager
	blockManager manager.BlockManager
	consensus    consensus.Consensus
	blockchain   *chain.BlockChain
	node         *node.Node
	server       *rpc.Server
}

func New(config *Config) *Trainer {
	t := &Trainer{
		self:    config.Self,
		config:  config,
		chainId: config.ChainId,
	}
	t.server = rpc.NewServer()
	t.blockchain = chain.NewBlockChain()
	t.config = config
	t.chainId = config.ChainId

	t.node = node.New(
		node.NewConfig(config.Self, config.BootstrapPeers, config.timeoutRPC, config.KeyImportOpts, config.HashOpts, config.SignerOpts),
		t.server)

	// register all the message listener to the node
	t.registerNodeListener()
	// begin to listen the messages
	go t.messageListener(t.server)
	go t.announceToNetwork()

	return t
}

func (t *Trainer) registerNodeListener() {
	// register the TrainerRegisterMessage
	t.node.RegisterListener(reflect.TypeOf(comm.TrainerRegisterMessage{}), message.NewTrainerRegisterListener(t.node.GetDiscovery()))
}

func (t *Trainer) messageListener(server *rpc.Server) {
	endpoint := t.self.Endpoint
	l, err := net.Listen("tcp", endpoint)
	if err != nil {
		logger.Panic("listen the endpoint failed")
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			logger.Panic("close the net listener failed")
		}
	}(l)

	for {
		conn, _ := l.Accept()
		go server.ServeConn(conn)
	}
}

func (t *Trainer) announceToNetwork() {
	peerRegisterMsg := &comm.Message{
		Content: &comm.TrainerRegisterMessage{Trainer: t.self},
		Header:  &proto.Header{ChainId: t.chainId, Timestamp: time.Now(), Creator: t.self.PublicKey},
	}
	peersToSend := make([]*comm.RemotePeer, len(t.config.BootstrapPeers))

	for _, peer := range t.config.BootstrapPeers {
		if !reflect.DeepEqual(peer, t.self) {
			peersToSend = append(peersToSend, peer)
			logger.Debug("will announce to peer:" + peer.Endpoint)
		}
	}

	t.node.SendToPeers(peerRegisterMsg, peersToSend)
}
