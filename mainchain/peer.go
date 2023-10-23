package mainchain

import (
	"encoding/gob"
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/block/chain"
	"github.com/wisecoach/ml_chain/block/consensus"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	consensus2 "github.com/wisecoach/ml_chain/mainchain/consensus"
	"github.com/wisecoach/ml_chain/mainchain/message"
	"github.com/wisecoach/ml_chain/mainchain/task"
	"github.com/wisecoach/ml_chain/mainchain/txhandler"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"net"
	"net/rpc"
	"reflect"
	"time"
)

type Peer struct {
	self    *proto.RemotePeer
	config  *Config
	chainId string

	taskClient   task.Client
	taskManager  task.Manager
	consensus    consensus.Consensus
	blockManager manager.BlockManager
	blockchain   *chain.BlockChain
	node         *node.Node
	server       *rpc.Server
	mcs          crypto.MessageCryptoService

	logger *zap.Logger
}

func New(config *Config) *Peer {
	p := &Peer{
		self:    config.Self,
		config:  config,
		logger:  log.GetLogger(config.Self.Endpoint),
		chainId: config.ChainId,
	}
	p.server = rpc.NewServer()
	p.blockchain = chain.NewBlockChain(&chain.Config{ChainId: config.ChainId, MaxBlockNumInMemory: config.MaxBlockNumInMemory})
	p.config = config

	// init mcs
	bccsp, err := sw.NewBCCSP()
	if err != nil {
		p.logger.Error("create bccsp failed: " + err.Error())
	}
	p.mcs = crypto.New(bccsp, config.Sk, config.Self, config.KeyImportOpts, config.HashOpts, config.SignerOpts)

	// init node
	p.node = node.New(
		node.NewConfig(config.Sk, config.Self, config.BootstrapPeers, config.Notaries, config.TimeoutRPC, config.KeyImportOpts, config.HashOpts, config.SignerOpts),
		p.server, p.mcs)

	// init the block manager
	p.blockManager = manager.New(p.blockchain, p.mcs)

	// init the consensus
	p.consensus = consensus2.New(&consensus2.Config{
		ChainId:           p.config.ChainId,
		HashInterval:      p.config.HashInterval,
		MaxInterval:       p.config.MaxInterval,
		MaxTxNum:          p.config.MaxTxNum,
		NumToConfirm:      p.config.NumToConfirm,
		DefaultDifficulty: p.config.DefaultDifficulty,
		GenesisBlock:      p.config.GenesisBlock,
	}, p.blockManager, p.node, p.mcs)

	// init task manager
	p.taskManager = task.NewTaskManager(p.config.Self)
	// init task client
	p.taskClient = task.NewTaskClient(&task.Config{
		ChainId: p.config.ChainId,
	}, p.taskManager, p.mcs, p.node)

	// register the message listener to the node
	p.registerNodeListener()
	// register block handler and tx handler
	p.registerBlockchainHandlers()

	// begin to listen the messages
	go p.messageListener(p.server)
	// wait for launch up
	select {
	case <-time.After(time.Second * 1):
	}
	go p.announceToNetwork()

	// wait for discovery
	select {
	case <-time.After(time.Second * 1):
	}
	p.logger.Info("init peer success, and begin to handle the task")

	// start consensus
	p.consensus.Start()
	// confirm genesis block to start the task
	err = p.blockManager.ConfirmBlock(p.config.GenesisBlock)
	if err != nil {
		p.logger.Error("confirm genesis block failed: " + err.Error())
		return nil
	}

	return p
}

func (p *Peer) registerNodeListener() {
	// register the PeerRegisterMessage
	p.node.RegisterListener(&proto.PeerRegisterMessage{}, message.NewPeerRegisterListener(p.node.GetDiscovery(), p.node))
	// register the BlockMessage and TransactionMessage
	p.node.RegisterListener(&proto.BlockMessage{}, message.NewBlockMessageListener(p.consensus))
	p.node.RegisterListener(&proto.TransactionMessage{}, message.NewTransactionMessageListener(p.consensus))
}

func (p *Peer) registerBlockchainHandlers() {
	// register the TaskGenesis transaction handler
	gob.Register(&proto.TaskGenesis{})
	p.blockManager.RegisterTxHandler(txhandler.NewTaskGenesisTxHandler(p.taskManager))
	p.blockManager.RegisterTxHandler(txhandler.NewTaskDeployManager(&txhandler.Config{
		TrainerNum:   p.config.TrainerNum,
		Sk:           p.config.Sk,
		ShardCreator: p.config.ShardCreator,
		ApiBaseUrl:   p.config.ApiBaseUrl,
	}, p.node, p.mcs))
	// register the TaskFinish transaction handler
	gob.Register(&proto.TaskResult{})
	p.blockManager.RegisterTxHandler(txhandler.NewTaskFinishTxHandler(p.taskManager))
	// register the ManagerRegister transaction handler
	gob.Register(&proto.ManagerRegister{})
	p.blockManager.RegisterTxHandler(txhandler.NewManagerRegisterTxHandler(p.taskManager))
	// register the ManagerRevoke transaction handler
	gob.Register(&proto.ManagerRevoke{})
	p.blockManager.RegisterTxHandler(txhandler.NewManagerRevokeTxHandler(p.taskManager))
}

func (p *Peer) messageListener(server *rpc.Server) {
	endpoint := p.self.Endpoint
	l, err := net.Listen("tcp", endpoint)
	p.logger.Info("begin to listen " + endpoint)
	if err != nil {
		p.logger.Panic("listen the endpoint failed")
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			p.logger.Panic("close the net listener failed")
		}
	}(l)

	for {
		conn, _ := l.Accept()
		go server.ServeConn(conn)
	}
}

func (p *Peer) announceToNetwork() {
	peerRegisterMsg := &proto.Message{
		Content: &proto.PeerRegisterMessage{Peer: p.self},
		Header:  &proto.Header{ChainId: p.chainId, Timestamp: time.Now(), Creator: p.self.PublicKey},
	}
	peersToSend := make([]*proto.RemotePeer, 0)

	for _, peer := range p.config.BootstrapPeers {
		if !reflect.DeepEqual(peer, p.self) {
			peersToSend = append(peersToSend, peer)
			p.logger.Debug("will announce to peer:" + peer.Endpoint)
		}
	}

	p.node.SendToPeers(peerRegisterMsg, peersToSend...)
}

func (p *Peer) CreateTask(task *task.Task) {
	p.taskClient.CreateTask(task)
}

func (p *Peer) FinishTask(task *task.FinishedTask) {
	p.taskClient.FinishTask(task)
}

func (p *Peer) RegisterManager(pk []byte) {
	p.taskClient.RegisterManager(pk)
}

func (p *Peer) RevokeManager(pk []byte) {
	p.taskClient.RevokeManager(pk)
}
