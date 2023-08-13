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
	"github.com/wisecoach/ml_chain/trainer/model/validate"
	"github.com/wisecoach/ml_chain/trainer/role"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"sync"
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
	validator   validate.Validator
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
		Self:            i.node.Self(),
		TaskManagerList: genesis.ManagerList,
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
		ValidatorNum: i.config.ValidatorNum,
		TaskId:       i.config.TaskId,
	}, i.mcs, i.roleManager, i, i.node, i.client)

	// init validator
	i.validator = validate.New(&validate.Config{
		TaskId: i.config.TaskId,
	}, i.client, i.mcs, i.node)

	// init aggregator
	i.aggregator = aggregate.New(&aggregate.Config{
		TaskId: i.config.TaskId,
	}, i.client, i.mcs, i.node, i, i.roleManager)

	// register message listener
	// trainer
	i.node.RegisterListener(&proto.ResponseLossMessage{}, message.NewValidateResponseMessageListener(i.trainer))
	// aggregator
	i.node.RegisterListener(&proto.SubmitLocalityWeightMessage{}, message.NewLocalWightMessageListener(i.aggregator))
	// validator
	i.node.RegisterListener(&proto.RequestLossMessage{}, message.NewValidateRequestMessageListener(i.validator))
	// consensus
	i.node.RegisterListener(&proto.BlockMessage{}, message.NewBlockMessageListener(i.consensus))
	gob.Register(&proto.TaskGenesis{})
	gob.Register(&proto.ModelIteration{})
	i.node.RegisterListener(&proto.TransactionMessage{}, message.NewTransactionMessageListener(i.consensus))

	// select validators
	latestBlock, err := i.blockManager.GetLatestBlock()
	if err != nil {
		i.logger.Error(err.Error())
	}
	i.roleManager.SelectValidators(i.node.Peers(), latestBlock.Header.DataHash, i.config.ValidatorNum)

	// start consensus
	go i.consensus.Start()

	// train new global weight
	go i.trainer.Train(genesis.InitWeight.Payload)

	// if self is aggregator, start aggregate
	if i.roleManager.AmAggregator(i.iteration) {
		go i.aggregator.StartAggregate()
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
	i.locker.Unlock()

	// select new validator for the new iteration
	latestBlock, err := i.blockManager.GetLatestBlock()
	if err != nil {
		i.logger.Error(err.Error())
	}

	i.roleManager.SelectValidators(i.node.Peers(), latestBlock.Header.DataHash, i.config.ValidatorNum)

	// load new global model in the new iteration
	go i.trainer.Train(iteration.GlobalWeight)

	if i.roleManager.AmAggregator(i.iteration) {
		go i.aggregator.StartAggregate()
	}
}
