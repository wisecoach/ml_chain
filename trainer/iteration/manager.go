package iteration

import (
	"encoding/gob"
	"fmt"
	"github.com/wisecoach/ml_chain/block/consensus"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	task_consensus "github.com/wisecoach/ml_chain/trainer/consensus"
	"github.com/wisecoach/ml_chain/trainer/message"
	"github.com/wisecoach/ml_chain/trainer/model/aggregate"
	"github.com/wisecoach/ml_chain/trainer/model/python"
	"github.com/wisecoach/ml_chain/trainer/model/train"
	"github.com/wisecoach/ml_chain/trainer/role"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Manager
// @Description: it's used for manage the iteration, and process for the new iteration
type Manager interface {
	// GetIteration
	//  @Description: get the current iteration
	//
	GetIteration() int

	// NextIteration
	//  @Description: start a new iteration, select new roles, load new global, train new local model
	//  @return error
	//
	NextIteration(iteration *proto.ModelIteration)

	// Start
	//  @Description: start first iteration by the task genesis, init role manager, consensus, trainer, aggregator, validator
	//  @param genesis
	//
	Start(genesis *proto.TaskGenesis)
}

// New
//
//	 @Description: create iteration manager, fill the selected fields, other like client, trainer, aggregator, validator,
//					  consensus will be initialized when blockMgr confirm the genesis block and call Start(genesis)
func New(config *Config, blockManager manager.BlockManager, node *node.Node, mcs crypto.MessageCryptoService) Manager {
	manager := &iterationManagerImpl{
		iteration:    0,
		locker:       sync.RWMutex{},
		logger:       log.GetLogger(node.Self().Endpoint),
		config:       config,
		blockManager: blockManager,
		node:         node,
		mcs:          mcs,
	}
	return manager
}

type iterationManagerImpl struct {
	iteration int
	locker    sync.RWMutex
	logger    *zap.Logger

	config      *Config
	client      python.Client
	trainer     train.LocalTrainer
	aggregator  aggregate.Aggregator
	roleManager role.Manager
	consensus   consensus.Consensus

	blockManager manager.BlockManager
	node         *node.Node
	mcs          crypto.MessageCryptoService
}

func (i *iterationManagerImpl) Start(genesis *proto.TaskGenesis) {
	i.locker.Lock()
	// init iteration, now global iteration is 0, so the iteration manager needs to iterate to 1
	i.iteration = 1
	i.locker.Unlock()

	i.logger.Info(fmt.Sprintf("start to manage iteration: %d", i.iteration))
	// init role manager
	i.roleManager = role.New(&role.Config{
		Self:       i.node.Self(),
		Aggregator: i.config.Aggregator,
	})

	// init consensus
	i.consensus = task_consensus.New(
		&task_consensus.Config{TaskId: i.config.TaskId},
		i.blockManager, i, i.roleManager, i.mcs, i.node)

	// init client
	i.client = python.New(&python.Config{
		ApiBaseUrl: i.config.ApiBaseUrl,
		TaskId:     i.config.TaskId,
	})

	// init trainer
	i.trainer = train.New(&train.Config{
		TaskId: i.config.TaskId,
		Cindex: i.config.Cindex,
	}, i.mcs, i.roleManager, i, i.node, i.client)

	// init aggregator
	i.aggregator = aggregate.New(&aggregate.Config{
		TaskId: i.config.TaskId,
	}, i.client, i.mcs, i.node, i, i.roleManager)

	// register message listener
	// aggregator
	i.node.RegisterListener(&proto.SubmitLocalityWeightMessage{}, message.NewLocalWightMessageListener(i.aggregator))
	// consensus
	i.node.RegisterListener(&proto.BlockMessage{}, message.NewBlockMessageListener(i.consensus))
	gob.Register(&proto.TaskGenesis{})
	gob.Register(&proto.ModelIteration{})
	i.node.RegisterListener(&proto.TransactionMessage{}, message.NewTransactionMessageListener(i.consensus))

	// wait for other trainer init iteration manager
	<-time.After(time.Second * 10)

	// init python server
	err := i.client.Init(genesis)
	if err != nil {
		i.logger.Error(err.Error())
		return
	}

	// start consensus
	go i.consensus.Start()

	// train new global weight
	go i.trainer.Train(genesis.InitWeight.Payload)

	// if self is aggregator, start aggregate
	if i.roleManager.AmAggregator(i.iteration) {
		go i.aggregator.StartAggregate(genesis.InitWeight.Payload)
	}
}

func (i *iterationManagerImpl) GetIteration() int {
	i.locker.RLock()
	defer i.locker.RUnlock()

	return i.iteration
}

func (i *iterationManagerImpl) NextIteration(iteration *proto.ModelIteration) {
	i.locker.Lock()
	// add the iteration
	i.iteration++
	i.logger.Info(fmt.Sprintf("start iteration %d", i.iteration))
	i.locker.Unlock()

	// load new global model in the new iteration
	go i.trainer.Train(iteration.GlobalWeight)

	if i.roleManager.AmAggregator(i.iteration) {
		go i.aggregator.StartAggregate(iteration.GlobalWeight)
	}
}
