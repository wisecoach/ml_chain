package message

import (
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/validate"
	"go.uber.org/zap"
)

type ValidateRequestMessageListener struct {
	logger    *zap.Logger
	Validator validate.Validator
}

func (v *ValidateRequestMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	req := message.Envelope.Payload.GetRequestLoss()
	if req == nil {
		v.logger.Error("message is not request LossRequest")
	}
	localWeight := &proto.LocalityWeight{
		Iteration:          req.Iteration,
		WeightVector:       req.WeightVector,
		Trainer:            req.Trainer,
		ValidatorSelection: req.ValidatorSelection,
		Losses:             nil,
	}
	v.Validator.Validate(localWeight)
}
