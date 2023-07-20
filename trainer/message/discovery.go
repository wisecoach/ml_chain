package message

import (
	"github.com/wisecoach/ml_chain/comm/discovery"
	"github.com/wisecoach/ml_chain/proto"
)

type TrainerRegisterListener struct {
	disc discovery.Discovery
}

func NewTrainerRegisterListener(disc discovery.Discovery) *TrainerRegisterListener {
	return &TrainerRegisterListener{
		disc: disc,
	}
}

func (t *TrainerRegisterListener) HandleMessage(message *proto.ReceivedMessage) {
	trainerRegister := message.Envelope.Payload.GetTrainerRegister()
	t.disc.Register(trainerRegister.Trainer)
}
