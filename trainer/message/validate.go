package message

import (
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/validate"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
)

type ValidateRequestMessageListener struct {
	logger    *zap.Logger
	validator validate.Validator
}

func NewValidateRequestMessageListener(validator validate.Validator) node.MessageListener {
	return &ValidateRequestMessageListener{
		logger:    log.GetLogger(),
		validator: validator,
	}
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
	v.validator.Validate(localWeight)
}
