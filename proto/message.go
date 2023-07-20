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

// SubmitLocalityWeightMessage
// @Description: message for trainer to submit its locality model to the aggregator
type SubmitLocalityWeightMessage struct {
	Iteration          int
	WeightVector       []float64
	Trainer            []byte
	ValidatorSelection *SelectionResult
}

// RequestLossMessage
// @Description: message for aggregator to request the loss，为什么不是trainer自己找validator
type RequestLossMessage struct {
	Iteration    int
	WeightVector []float64
	Trainer      []byte
}

// ResponseLossMessage
// @Description: message for validator to response loss to the aggregator
type ResponseLossMessage struct {
	Iteration int
	Trainer   []byte
	Losses    []*Envelope[*ValidateLoss]
}

func (t *TrainerRegisterMessage) isMessage_Content() {}

func (m *Message) GetTrainerRegister() *TrainerRegisterMessage {
	if msg, ok := m.Content.(*TrainerRegisterMessage); ok {
		return msg
	}
	return nil
}

func (s *SubmitLocalityWeightMessage) isMessage_Content() {}

func (m *Message) GetSubmitLocalityWeight() *SubmitLocalityWeightMessage {
	if msg, ok := m.Content.(*SubmitLocalityWeightMessage); ok {
		return msg
	}
	return nil
}

func (s *RequestLossMessage) isMessage_Content() {}

func (m *Message) GetRequestLoss() *RequestLossMessage {
	if msg, ok := m.Content.(*RequestLossMessage); ok {
		return msg
	}
	return nil
}

func (s *ResponseLossMessage) isMessage_Content() {}

func (m *Message) GetResponseLoss() *ResponseLossMessage {
	if msg, ok := m.Content.(*ResponseLossMessage); ok {
		return msg
	}
	return nil
}
