package message

import (
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/train"
)

type ValidateResponseMessageListener struct {
	trainer train.LocalTrainer
}

func NewValidateResponseMessageListener(trainer train.LocalTrainer) node.MessageListener {
	return &ValidateResponseMessageListener{
		trainer: trainer,
	}
}

func (v *ValidateResponseMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	responseLoss := message.Envelope.Payload.GetResponseLoss()
	loss := responseLoss.Loss
	go v.trainer.CollectionLoss(loss)
}
