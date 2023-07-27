package train

import (
	"fmt"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/iteration"
	"github.com/wisecoach/ml_chain/trainer/model/python"
	"github.com/wisecoach/ml_chain/trainer/role"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"time"
)

type localTrainerImpl struct {
	taskId           string
	mcs              crypto.MessageCryptoService
	self             *comm.RemotePeer
	client           python.Client
	roleManager      role.Manager
	iterationManager iteration.Manager
	logger           zap.Logger
	node             *node.Node
	config           *Config

	lock              sync.RWMutex
	localModel        *proto.LocalityWeight
	validateWaitGroup sync.WaitGroup
}

func (l *localTrainerImpl) Train(weight *proto.GlobalWeight) {

	// train the global model
	l.logger.Info(fmt.Sprintf("begin to train the global model, iteration: %d", weight.Iteration))
	trainResponse, err := l.client.Train(&python.TrainRequest{GlobalModel: weight})
	if err != nil {
		return
	}
	l.logger.Info(fmt.Sprintf("trainnging finished, %d", trainResponse.LocalModel.Iteration))
	l.lock.Lock()
	l.localModel = trainResponse.LocalModel
	l.localModel.Iteration = l.iterationManager.GetIteration()
	l.localModel.Losses = make([]*proto.Envelope[*proto.ValidateLoss], l.config.ValidatorNum)

	// get the validator selected by iterationManager
	validatorSelection := l.roleManager.GetValidators()

	// send request to the validators, and wait for the validate response
	l.validateWaitGroup.Add(l.config.ValidatorNum)
	l.logger.Info(fmt.Sprintf("begin to send loss request to %d validators", l.config.ValidatorNum))
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
	defer l.lock.Unlock()
	l.logger.Info(fmt.Sprintf("have colllected %d validate response, begin to send to the aggregator", l.config.ValidatorNum))

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
	}

	l.node.SendToPeers(localModelSubmitMsg, aggregator)
	l.localModel = nil

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
