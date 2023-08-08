package proto

type RemotePeer struct {
	Endpoint  string
	PublicKey []byte
}

type ReceivedMessage struct {
	Envelope *Envelope[*Message]
	Sender   *RemotePeer
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

// TransactionMessage
// @Description: message for propose transaction to consensus
type TransactionMessage struct {
	Transaction *Envelope[*Transaction]
}

// BlockMessage
// @Description: message for deliver block
type BlockMessage struct {
	Block *Envelope[*Block]
}

// PeerRegisterMessage
// @Description: message for discovery, register self to the network
type PeerRegisterMessage struct {
	Peer *RemotePeer
}

// SubmitLocalityWeightMessage
// @Description: message for trainer to submit its locality model to the aggregator
type SubmitLocalityWeightMessage struct {
	LocalityWeight *LocalityWeight
}

// RequestLossMessage
// @Description: message for aggregator to request the loss，为什么不是trainer自己找validator
type RequestLossMessage struct {
	Iteration          int
	WeightVector       []float64
	Trainer            []byte
	ValidatorSelection *SelectionResult
}

// ResponseLossMessage
// @Description: message for validator to response loss to the aggregator
type ResponseLossMessage struct {
	Loss *Envelope[*ValidateLoss]
}

func (t *TransactionMessage) isMessage_Content() {}

func (m *Message) GetTransaction() *TransactionMessage {
	if msg, ok := m.Content.(*TransactionMessage); ok {
		return msg
	}
	return nil
}

func (b *BlockMessage) isMessage_Content() {}

func (m *Message) GetBlock() *BlockMessage {
	if msg, ok := m.Content.(*BlockMessage); ok {
		return msg
	}
	return nil
}

func (p *PeerRegisterMessage) isMessage_Content() {}

func (m *Message) GetPeerRegitser() *PeerRegisterMessage {
	if msg, ok := m.Content.(*PeerRegisterMessage); ok {
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
