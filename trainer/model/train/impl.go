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
	"reflect"
	"sync"
	"time"
)

type IterationMgrAdapter interface {
	GetIteration() int
}

type localTrainerImpl struct {
	taskId           string
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

	trainResponse, err := l.client.Train(&python.TrainRequest{TrainerId: l.self.Endpoint, GlobalModel: weight})
	if err != nil {
		return
	}
	//trainResponse := &python.TrainResponse{
	//	LocalModel: &proto.LocalityWeight{
	//		Iteration:          0,
	//		WeightVector:       nil,
	//		Trainer:            l.self.PublicKey,
	//		ValidatorSelection: nil,
	//		Losses:             nil,
	//	},
	//}
	// select {
	// case <-time.After(time.Millisecond * 1):
	// }

	l.logger.Info(fmt.Sprintf("trainnging finished, %d", trainResponse.LocalModel.Iteration))
	l.lock.Lock()
	l.localModel = trainResponse.LocalModel
	l.localModel.Trainer = l.self.PublicKey
	l.localModel.Iteration = l.iterationManager.GetIteration()
	l.localModel.Losses = make([]*proto.Envelope[*proto.ValidateLoss], l.config.ValidatorNum)

	// get the validator selected by iterationManager
	validatorSelection := l.roleManager.GetValidators()
	l.localModel.ValidatorSelection = validatorSelection

	// send request to the validators, and wait for the validate response
	l.validateWaitGroup.Add(l.config.ValidatorNum)
	logStr := fmt.Sprintf("begin to send loss request to %d validators: ", l.config.ValidatorNum)
	for _, winner := range validatorSelection.Winners {
		logStr += winner.Endpoint + ","
	}
	l.logger.Info(logStr)
	lossReqMsg := &proto.Message{
		Content: &proto.RequestLossMessage{
			Iteration:          l.localModel.Iteration,
			WeightVector:       l.localModel.WeightVector,
			Trainer:            l.localModel.Trainer,
			ValidatorSelection: validatorSelection,
		},
		Header: &proto.Header{
			Creator:   l.self.PublicKey,
			ChainId:   l.taskId,
			TxId:      "",
			Timestamp: time.Now(),
		},
	}
	l.lock.Unlock()

	l.node.SendToPeers(lossReqMsg, validatorSelection.Winners...)
	l.logger.Info(fmt.Sprintf("begin to wait for %d loss response", l.config.ValidatorNum))
	l.validateWaitGroup.Wait()

	l.lock.Lock()

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

	l.logger.Info(fmt.Sprintf("have collected %d validate response, begin to send to the aggregator: %s",
		l.config.ValidatorNum, aggregator.Endpoint))

	l.node.SendToPeers(localModelSubmitMsg, aggregator)
	l.localModel = nil

	l.lock.Unlock()
}

func (l *localTrainerImpl) CollectionLoss(loss *proto.Envelope[*proto.ValidateLoss]) {
	l.lock.Lock()
	defer l.lock.Unlock()

	losses := l.localModel.Losses
	// make loss's index is equal to validator's index in vrf selection
	index := -1
	for i, validator := range l.localModel.ValidatorSelection.Winners {
		if reflect.DeepEqual(validator.PublicKey, loss.Payload.Validator) {
			index = i
		}
	}
	if index == -1 {
		l.logger.Error("the validator of loss is not in vrf selection")
		return
	}
	losses[index] = loss
	l.localModel.Losses = losses
	l.validateWaitGroup.Done()
}
