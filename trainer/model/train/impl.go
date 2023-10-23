package train

import (
	"fmt"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/python"
	"github.com/wisecoach/ml_chain/trainer/role"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"sync"
	"time"
)

type IterationMgrAdapter interface {
	GetIteration() int
}

type localTrainerImpl struct {
	taskId           string
	cindex           string
	mcs              crypto.MessageCryptoService
	self             *proto.RemotePeer
	client           python.Client
	roleManager      role.Manager
	iterationManager IterationMgrAdapter
	logger           *zap.Logger
	node             *node.Node
	config           *Config

	lock              sync.RWMutex
	localModel        *proto.LocalityWeight
	validateWaitGroup sync.WaitGroup
}

func New(config *Config, mcs crypto.MessageCryptoService, roleManager role.Manager, iterationManager IterationMgrAdapter,
	node *node.Node, client python.Client) LocalTrainer {
	trainer := &localTrainerImpl{
		taskId:            config.TaskId,
		cindex:            config.Cindex,
		mcs:               mcs,
		self:              node.Self(),
		client:            client,
		roleManager:       roleManager,
		iterationManager:  iterationManager,
		logger:            log.GetLogger(node.Self().Endpoint),
		node:              node,
		config:            config,
		lock:              sync.RWMutex{},
		validateWaitGroup: sync.WaitGroup{},
	}
	return trainer
}

func (l *localTrainerImpl) Train(weight *proto.GlobalWeight) {

	// train the global model
	l.logger.Info(fmt.Sprintf("begin to train the global model, iteration: %d", weight.Iteration))

	trainResponse, err := l.client.Train(&python.TrainRequest{
		Cindex:      l.cindex,
		Iteration:   l.iterationManager.GetIteration(),
		GlobalModel: weight,
	})
	if err != nil {
		l.logger.Error("train failed, err: " + err.Error())
		return
	}
	localModel := trainResponse.LocalModel
	localModel.Cindex = l.cindex
	localModel.Iteration = l.iterationManager.GetIteration()
	l.logger.Info(fmt.Sprintf("trainnging finished, iteration = %d, n_samples = %d, model_hash = %s",
		localModel.Iteration, localModel.NSamples, localModel.ModelHash))

	l.lock.Lock()
	l.localModel = localModel

	localModelSubmitMsg := &proto.Message{
		Content: &proto.SubmitLocalityWeightMessage{
			LocalityWeight: l.localModel,
		},
		Header: &proto.Header{
			Creator:   l.self.PublicKey,
			ChainId:   l.taskId,
			TxId:      "",
			Timestamp: time.Now(),
		},
	}

	// 选取 聚合者
	aggregatorPk := l.roleManager.GetAggregator(l.iterationManager.GetIteration())
	aggregator := l.node.Lookup(aggregatorPk)
	if aggregator == nil {
		l.logger.Error("cannot found aggregator")
		fmt.Println("cannot found aggregator, self: " + l.self.Endpoint)
	}

	l.logger.Debug(fmt.Sprintf("have collected %d validate response, begin to send to the aggregator: %s",
		l.config.ValidatorNum, aggregator.Endpoint))

	l.node.SendToPeers(localModelSubmitMsg, aggregator)
	l.localModel = nil

	l.lock.Unlock()
}
