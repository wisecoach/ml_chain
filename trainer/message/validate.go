package message

import (
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/validate"
)

type ValidateRequestMessageListener struct {
	validator validate.Validator
}

func NewValidateRequestMessageListener(validator validate.Validator) comm.MessageListener {
	return &ValidateRequestMessageListener{
		validator: validator,
	}
}

func (v *ValidateRequestMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	req := message.Envelope.Payload.GetRequestLoss()
	localWeight := &proto.LocalityWeight{
		Iteration:          req.Iteration,
		WeightVector:       req.WeightVector,
		Trainer:            req.Trainer,
		ValidatorSelection: req.ValidatorSelection,
		Losses:             nil,
	}
	v.validator.Validate(localWeight)
}
