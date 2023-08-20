package trainer

import (
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/block/chain"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/iteration"
	"github.com/wisecoach/ml_chain/trainer/message"
	"github.com/wisecoach/ml_chain/trainer/txhandler"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"net"
	"net/rpc"
	"reflect"
	"time"
)

type Trainer struct {
	self   *proto.RemotePeer
	config *Config
	taskId string

	iterationManager iteration.Manager
	blockManager     manager.BlockManager
	blockchain       *chain.BlockChain
	node             *node.Node
	server           *rpc.Server
	mcs              crypto.MessageCryptoService

	logger *zap.Logger
}

func New(config *Config) *Trainer {
	t := &Trainer{
		self:   config.Self,
		config: config,
		taskId: config.TaskId,
		logger: log.GetLogger(config.Self.Endpoint),
	}
	t.server = rpc.NewServer()
	t.blockchain = chain.NewBlockChain()
	t.config = config
	t.taskId = config.TaskId

	// init mcs
	bccsp, err := sw.NewBCCSP()
	if err != nil {
		t.logger.Error("create bccsp failed: " + err.Error())
	}
	t.mcs = crypto.New(bccsp, config.Sk, config.Self, config.KeyImportOpts, config.HashOpts, config.SignerOpts)

	// init node
	t.node = node.New(
		node.NewConfig(config.Sk, config.Self, config.BootstrapPeers, config.Notaries, config.TimeoutRPC, config.KeyImportOpts, config.HashOpts, config.SignerOpts),
		t.server, t.mcs)

	// init the block manager
	t.blockManager = manager.New(t.blockchain)

	// init the iteration manager
	t.iterationManager = iteration.New(&iteration.Config{
		ValidatorNum: t.config.ValidatorNum,
		TaskId:       t.config.TaskId,
		ApiBaseUrl:   t.config.ApiBaseUrl,
	}, t.blockManager, t.node, t.mcs)

	// register the message listener to the node
	t.registerNodeListener()
	// register block handler and tx handler
	t.registerBlockchainHandlers()

	// begin to listen the messages
	go t.messageListener(t.server)
	// wait for launch up
	select {
	case <-time.After(time.Second * 1):
	}
	go t.announceToNetwork()

	// wait for discovery
	select {
	case <-time.After(time.Second * 1):
	}
	t.logger.Info("init trainer success, and begin to handle the task")

	// confirm genesis block to start the task
	err = t.blockManager.ConfirmBlock(t.config.GenesisBlock)
	if err != nil {
		t.logger.Error("confirm genesis block failed: " + err.Error())
		return nil
	}

	return t
}

func (t *Trainer) registerNodeListener() {
	// register the PeerRegisterMessage
	t.node.RegisterListener(&proto.PeerRegisterMessage{}, message.NewPeerRegisterListener(t.node.GetDiscovery(), t.node))
}

func (t *Trainer) registerBlockchainHandlers() {
	// register the TaskGenesis transaction handler
	t.blockManager.RegisterTxHandler(txhandler.NewGenesisTxHandler(t.iterationManager))
	// register the ModelIteration transaction handler
	t.blockManager.RegisterTxHandler(txhandler.NewIterationTxHandler(t.iterationManager))
}

func (t *Trainer) messageListener(server *rpc.Server) {
	endpoint := t.self.Endpoint
	l, err := net.Listen("tcp", endpoint)
	t.logger.Info("begin to listen " + endpoint)
	if err != nil {
		t.logger.Panic("listen the endpoint failed")
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			t.logger.Panic("close the net listener failed")
		}
	}(l)

	for {
		conn, _ := l.Accept()
		go server.ServeConn(conn)
	}
}

func (t *Trainer) announceToNetwork() {
	peerRegisterMsg := &proto.Message{
		Content: &proto.PeerRegisterMessage{Peer: t.self},
		Header:  &proto.Header{ChainId: t.taskId, Timestamp: time.Now(), Creator: t.self.PublicKey},
	}
	peersToSend := make([]*proto.RemotePeer, 0)

	for _, peer := range t.config.BootstrapPeers {
		if !reflect.DeepEqual(peer, t.self) {
			peersToSend = append(peersToSend, peer)
			t.logger.Debug("will announce to peer:" + peer.Endpoint)
		}
	}

	t.node.SendToPeers(peerRegisterMsg, peersToSend...)
}
