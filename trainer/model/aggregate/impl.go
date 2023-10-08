package aggregate

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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

type aggregatorImpl struct {
	lock             sync.Mutex
	trainerWaitGroup sync.WaitGroup
	waitAggregate    sync.WaitGroup
	logger           *zap.Logger
	config           *Config

	client           python.Client
	mcs              crypto.MessageCryptoService
	node             *node.Node
	iterationManager IterationMgrAdapter
	roleManager      role.Manager
	localModels      map[string]*proto.LocalityWeight
}

func New(config *Config, client python.Client, mcs crypto.MessageCryptoService, node *node.Node,
	iterationManager IterationMgrAdapter, roleManager role.Manager) Aggregator {
	a := &aggregatorImpl{
		lock:             sync.Mutex{},
		trainerWaitGroup: sync.WaitGroup{},
		logger:           log.GetLogger(node.Self().Endpoint),
		config:           config,
		client:           client,
		mcs:              mcs,
		node:             node,
		iterationManager: iterationManager,
		roleManager:      roleManager,
		localModels:      make(map[string]*proto.LocalityWeight),
	}
	a.waitAggregate.Add(1)
	return a
}

func (a *aggregatorImpl) HandleLocalModel(weight *proto.LocalityWeight) {
	// wait aggregate process to start, note: wait before lock
	a.waitAggregate.Wait()

	a.lock.Lock()
	defer a.lock.Unlock()

	a.logger.Debug(fmt.Sprintf("have received %d local model, %s receive local model, iteration: %d, from: %s",
		len(a.localModels), a.node.Self().Endpoint, weight.Iteration, a.node.Lookup(weight.Trainer).Endpoint))
	// now, we don't receive new model from same trainer
	_, exist := a.localModels[string(weight.Trainer)]
	if exist {
		a.logger.Error("receive duplicated model")
		return
	}
	if a.iterationManager.GetIteration() > weight.Iteration {
		a.logger.Error("receive expired model")
		return
	}
	if a.iterationManager.GetIteration() != weight.Iteration {
		a.logger.Error(fmt.Sprintf("receive model %d whose global model newer than us %d", weight.Iteration, a.iterationManager.GetIteration()))
		return
	}
	// validate the local model
	err := a.validateLocalModel(weight)
	if err != nil {
		a.logger.Error("local model is invalid: " + err.Error())
		return
	}
	a.trainerWaitGroup.Done()
	// set the local model
	a.localModels[string(weight.Trainer)] = weight
}

func (a *aggregatorImpl) StartAggregate() {
	peers := a.node.Peers()
	trainerNum := len(peers)
	a.logger.Info(fmt.Sprintf("wait for %d trainer's local model to aggregate", trainerNum))
	a.trainerWaitGroup.Add(trainerNum)
	a.waitAggregate.Done()
	a.trainerWaitGroup.Wait()
	a.waitAggregate.Add(1)
	a.logger.Info(fmt.Sprintf("aggregate global model begin: iteration = %d", a.iterationManager.GetIteration()))

	a.lock.Lock()
	localModels := make([]*proto.LocalityWeight, 0)
	for _, value := range a.localModels {
		localModels = append(localModels, value)
	}
	a.lock.Unlock()

	// aggregate the local model by python client
	request := &python.AggregateRequest{LocalModels: localModels}
	response, err := a.client.Aggregate(request)
	if err != nil {
		a.logger.Error("python aggregate failed: " + err.Error())
		return
	}

	// package the global model to transaction, send to consensus model
	globalModel := response.GlobalModel
	globalModel.Aggregator = a.node.Self().PublicKey
	globalModel.Iteration = a.iterationManager.GetIteration()

	a.logger.Info(fmt.Sprintf("aggregate global model success: iteration: %d, acc: %f, loss: %f",
		globalModel.Iteration, globalModel.TotalAcc, globalModel.Loss))

	transaction := &proto.Transaction{
		Parent: nil,
		Header: &proto.Header{
			Creator:   a.node.Self().PublicKey,
			ChainId:   a.config.TaskId,
			TxId:      uuid.New().String(),
			Timestamp: time.Time{},
		},
		Payload: &proto.ModelIteration{
			Iteration:    a.iterationManager.GetIteration(),
			GlobalWeight: globalModel,
			// LocalityWeights: localModels,
		},
	}
	txBytes, err := json.Marshal(transaction)
	if err != nil {
		a.logger.Error("transaction marshal failed")
		return
	}
	signature, err := a.mcs.Sign(txBytes)
	if err != nil {
		a.logger.Error("transaction sign failed")
		return
	}
	signedTransaction := &proto.Envelope[*proto.Transaction]{
		Payload:   transaction,
		Signature: signature,
	}
	msg := &proto.Message{
		Content: &proto.TransactionMessage{
			Transaction: signedTransaction,
		},
		Header: &proto.Header{
			Creator:   a.node.Self().PublicKey,
			ChainId:   a.config.TaskId,
			TxId:      "",
			Timestamp: time.Time{},
		},
	}
	a.node.SendToPeers(msg, a.node.Self())

	// clear the local model
	a.lock.Lock()
	a.localModels = make(map[string]*proto.LocalityWeight)
	a.lock.Unlock()
}

func (a *aggregatorImpl) validateLocalModel(model *proto.LocalityWeight) error {
	// for i, loss := range model.Losses {
	// 	// validate the signature
	// 	lossBytes, err := json.Marshal(loss.Payload)
	// 	if err != nil {
	// 		return errors.New("json marshal failed: " + err.Error())
	// 	}
	// 	valid, err := a.mcs.Verify(loss.Payload.Validator, lossBytes, loss.Signature)
	// 	if err != nil || !valid {
	// 		return errors.New("signature is invalid: " + err.Error())
	// 	}
	//
	// 	// validate vrf
	// 	err = a.roleManager.VerifyValidatorSelection(model.ValidatorSelection)
	// 	if err != nil {
	// 		return errors.New("vrf is invalid: " + err.Error())
	// 	}
	// 	if reflect.DeepEqual(loss.Payload.Validator, model.ValidatorSelection.Winners[i]) {
	// 		return errors.New("validator is not selected one")
	// 	}
	// }
	return nil
}
