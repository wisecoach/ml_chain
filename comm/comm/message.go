package comm

import "github.com/wisecoach/ml_chain/proto"

type ReceivedMessage struct {
	Envelope *proto.Envelope[*Message]
	Sender   *RemotePeer
}

// Message define the message sent in the network
type Message struct {
	Content isMessage_Content
	Header  *proto.Header
}

// isMessage_Content is the interface of the message content
type isMessage_Content interface {
	isMessage_Content()
}

// TrainerRegisterMessage
// @Description: message for discovery, register self to the network
type TrainerRegisterMessage struct {
	Trainer *RemotePeer
}

func (t *TrainerRegisterMessage) isMessage_Content() {}

func (m *Message) GetTrainerRegister() *TrainerRegisterMessage {
	if msg, ok := m.Content.(*TrainerRegisterMessage); ok {
		return msg
	}
	return nil
}
