package message

import (
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/train"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
)

type ValidateResponseMessageListener struct {
	logger  *zap.Logger
	trainer train.LocalTrainer
}

func NewValidateResponseMessageListener(trainer train.LocalTrainer) node.MessageListener {
	return &ValidateResponseMessageListener{
		logger:  log.GetLogger(),
		trainer: trainer,
	}
}

func (v *ValidateResponseMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	responseLoss := message.Envelope.Payload.GetResponseLoss()
	if responseLoss == nil {
		v.logger.Error("loss is nil")
	}
	loss := responseLoss.Loss
	go v.trainer.CollectionLoss(loss)
}
