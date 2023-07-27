package iteration

import (
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/aggregate"
	"github.com/wisecoach/ml_chain/trainer/model/train"
	"github.com/wisecoach/ml_chain/trainer/role"
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
}

type iterationManagerImpl struct {
	iteration int
	locker    sync.RWMutex
	logger    zap.Logger

	config       *Config
	trainer      train.LocalTrainer
	aggregator   aggregate.Aggregator
	roleManager  role.Manager
	blockManager manager.BlockManager
	node         *node.Node
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
	i.locker.Lock()

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
