package message

import (
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/train"
	"go.uber.org/zap"
)

type ValidateResponseMessageListener struct {
	logger  zap.Logger
	trainer train.LocalTrainer
}

func (v *ValidateResponseMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	responseLoss := message.Envelope.Payload.GetResponseLoss()
	if responseLoss == nil {
		v.logger.Error("loss is nil")
	}
	loss := responseLoss.Loss
	go v.trainer.CollectionLoss(loss)
}
