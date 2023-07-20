package comm

import "github.com/wisecoach/ml_chain/proto"

// Comm is an object that enables to communicate with other peers
// that also embed a CommModule.
type Comm interface {
	Self() *RemotePeer

	// Send sends a message to remote peers asynchronously
	Send(msg *proto.Envelope[*proto.Message], peers ...*RemotePeer)

	// Accept returns a dedicated read-only channel for messages sent by other nodes that match a certain predicate.
	// Each message from the channel can be used to send a reply back to the sender
	Accept(MessageAcceptor) <-chan *proto.ReceivedMessage

	// HandleMessage
	//  @Description: handle the message from rpc-server
	HandleMessage(message *proto.ReceivedMessage)

	// Stop stops the module
	Stop()
}

type RemotePeer struct {
	Endpoint  string
	PublicKey []byte
}

// MessageAcceptor 判断是否接收该消息的函数
type MessageAcceptor func(message *proto.ReceivedMessage) bool
