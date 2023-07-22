package validate

import (
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/python"
	"go.uber.org/zap"
)

type validatorImpl struct {
	self   *comm.RemotePeer
	client python.Client
	logger zap.Logger
	node   *node.Node
}

func (v validatorImpl) Validate(weight *proto.LocalityWeight) {
	lossResp, err := v.client.Validate(&python.ValidateRequest{Model: weight})
	if err != nil {
		v.logger.Error("cannot validate the model from: " + string(weight.Trainer))
	}
	loss := lossResp.Loss
	trainer := v.node.Lookup(weight.Trainer)
	msg := &proto.Message{
		Content: &proto.ResponseLossMessage{
			Iteration: weight.Iteration,
			Trainer:   trainer.PublicKey,
			Loss:      loss,
		},
		Header:  nil,
	}

	v.node.SendToPeers(, trainer)

}
