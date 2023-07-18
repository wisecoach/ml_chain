package message

import (
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/discovery"
)

type TrainerRegisterListener struct {
	disc discovery.Discovery
}

func NewTrainerRegisterListener(disc discovery.Discovery) *TrainerRegisterListener {
	return &TrainerRegisterListener{
		disc: disc,
	}
}

func (t *TrainerRegisterListener) HandleMessage(message *comm.ReceivedMessage) {
	trainerRegister := message.Envelope.Payload.GetTrainerRegister()
	t.disc.Register(trainerRegister.Trainer)
}
