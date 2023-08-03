package validate

import (
	"encoding/json"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/python"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
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
		logger: log.GetLogger(),
		self:   node.Self(),
		config: config,
		client: client,
		mcs:    mcs,
		node:   node,
	}
	return v
}

func (v validatorImpl) Validate(weight *proto.LocalityWeight) {
	lossResp, err := v.client.Validate(&python.ValidateRequest{Model: weight})
	if err != nil {
		v.logger.Error("cannot validate the model from: " + string(weight.Trainer))
		return
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

	v.node.SendToPeers(msg, trainer)
}
