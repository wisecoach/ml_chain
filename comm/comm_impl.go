package comm

import (
	"github.com/pkg/errors"
	"net/rpc"
)

// RPCPeer it's used for go RPC
type RPCPeer struct {
	handler func(interface{})
}

func (r *RPCPeer) ReadMessage(message Message) {
	r.handler(message)
}

type commImpl struct {
	*ChannelDeMultiplexer
}

func (c *commImpl) GetPublicKey() []byte {
	// TODO implement me
	panic("implement me")
}

func (c *commImpl) Send(msg *SignedMessage, peers ...*RemotePeer) {
	// TODO implement me
	panic("implement me")
}

func (c *commImpl) Accept(acceptor MessageAcceptor) <-chan SignedMessage {
	// TODO implement me
	panic("implement me")
}

func (c *commImpl) Stop() {
	// TODO implement me
	panic("implement me")
}

func New(server *rpc.Server) Comm {
	commInst := &commImpl{
		ChannelDeMultiplexer: NewChannelDemultiplexer(),
	}
	handler := func(message interface{}) {
		commInst.DeMultiplex(message)
	}

	rpcPeer := &RPCPeer{handler: handler}
	err := server.Register(rpcPeer)
	if err != nil {
		errors.Errorf("注册服务失败")
	}
	return commInst
}
