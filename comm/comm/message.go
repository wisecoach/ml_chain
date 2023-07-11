package comm

type ReceivedMessage struct {
	*SignedMessage
	sender *RemotePeer
}

type SignedMessage struct {
	*Message
	Signature []byte
}

// Message define the message sent in the network
type Message struct {
	Content isMessage_Content
}

// isMessage_Content is the interface of the message content
type isMessage_Content interface {
	isMessage_Content()
}
