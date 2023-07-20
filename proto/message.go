package proto

import "github.com/wisecoach/ml_chain/comm/comm"

type ReceivedMessage struct {
	Envelope *Envelope[*Message]
	Sender   *comm.RemotePeer
}

// Message define the message sent in the network
type Message struct {
	Content isMessage_Content
	Header  *Header
}

// isMessage_Content is the interface of the message content
type isMessage_Content interface {
	isMessage_Content()
}

// TrainerRegisterMessage
// @Description: message for discovery, register self to the network
type TrainerRegisterMessage struct {
	Trainer *comm.RemotePeer
}

func (t *TrainerRegisterMessage) isMessage_Content() {}

func (m *Message) GetTrainerRegister() *TrainerRegisterMessage {
	if msg, ok := m.Content.(*TrainerRegisterMessage); ok {
		return msg
	}
	return nil
}
