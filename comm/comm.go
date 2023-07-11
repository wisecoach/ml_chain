package comm

// Comm is an object that enables to communicate with other peers
// that also embed a CommModule.
type Comm interface {
	// GetPublicKey
	//  @Description: get the public key of the peer
	//  @return []byte
	GetPublicKey() []byte

	// Send sends a message to remote peers asynchronously
	Send(msg *SignedMessage, peers ...*RemotePeer)

	// Accept returns a dedicated read-only channel for messages sent by other nodes that match a certain predicate.
	// Each message from the channel can be used to send a reply back to the sender
	Accept(MessageAcceptor) <-chan SignedMessage

	DeMultiplex(msg interface{})

	// Stop stops the module
	Stop()
}

type RemotePeer struct {
	Endpoint  string
	PublicKey []byte
}

// MessageAcceptor 判断是否接收该消息的函数
type MessageAcceptor func(interface{}) bool
