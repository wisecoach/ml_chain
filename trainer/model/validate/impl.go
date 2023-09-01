package validate

import (
	"encoding/json"
	"fmt"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/python"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type validatorImpl struct {
	TaskId string
	logger *zap.Logger
	self   *proto.RemotePeer
	config *Config

	client python.Client
	mcs    crypto.MessageCryptoService
	node   *node.Node
}

func New(config *Config, client python.Client, mcs crypto.MessageCryptoService, node *node.Node) Validator {
	v := &validatorImpl{
		TaskId: config.TaskId,
		logger: log.GetLogger(node.Self().Endpoint),
		self:   node.Self(),
		config: config,
		client: client,
		mcs:    mcs,
		node:   node,
	}
	return v
}

func (v *validatorImpl) Validate(weight *proto.LocalityWeight) {
	v.logger.Debug(fmt.Sprintf("begin to validate the local model, iteration: %d", weight.Iteration))
	isFirst := reflect.DeepEqual(weight.ValidatorSelection.Winners[0].PublicKey, v.self.PublicKey)
	if isFirst {
		v.logger.Info(fmt.Sprintf("begin to validate the local model, iteration: %d", weight.Iteration))
	}
	lossResp, err := v.client.Validate(&python.ValidateRequest{Model: weight})
	if err != nil {
		v.logger.Error("cannot validate the model from: " + string(weight.Trainer) + ", " + err.Error())
		return
	}
	if isFirst {
		v.logger.Info(fmt.Sprintf("validate the local model success: iteration = %d, acc = %f, loss = %f",
			weight.Iteration, lossResp.Loss.Acc, lossResp.Loss.Loss))
	}

	loss := lossResp.Loss
	loss.Validator = v.self.PublicKey

	modelBytes, err := json.Marshal(weight)
	if err != nil {
		v.logger.Error("json marshal model failed:" + err.Error())
		return
	}
	loss.ModelHash, err = v.mcs.Hash(modelBytes)
	if err != nil {
		v.logger.Error("hash model failed:" + err.Error())
		return
	}

	trainer := v.node.Lookup(weight.Trainer)

	// sign the Loss
	lossBytes, err := json.Marshal(loss)
	if err != nil {
		v.logger.Error("loss json marshal failed:" + err.Error())
		return
	}
	signature, err := v.mcs.Sign(lossBytes)
	if err != nil {
		v.logger.Error("loss sign failed:" + err.Error())
		return
	}

	envelope := &proto.Envelope[*proto.ValidateLoss]{
		Signature: signature,
		Payload:   loss,
	}

	msg := &proto.Message{
		Content: &proto.ResponseLossMessage{
			Loss: envelope,
		},
		Header: &proto.Header{
			Creator:   v.self.PublicKey,
			ChainId:   v.TaskId,
			TxId:      "",
			Timestamp: time.Now(),
		},
	}

	v.logger.Debug(fmt.Sprintf("validating finished, %d, send to trainer: %s", weight.Iteration, trainer.Endpoint))
	v.node.SendToPeers(msg, trainer)
}
