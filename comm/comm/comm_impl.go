package comm

import (
	"fmt"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"
)

type MessageHandler struct {
	comm Comm
}

type subscription struct {
	ch   chan *proto.ReceivedMessage
	pred MessageAcceptor
}

// HandleMessage
//
//	@Description:
//	@receiver mh
//	@param message
//	@param reply	it's always be nil, and shouldn't be operated
func (mh *MessageHandler) HandleMessage(message *proto.ReceivedMessage, reply *proto.ReceivedMessage) error {
	mh.comm.HandleMessage(message)
	return nil
}

func New(server *rpc.Server, self *proto.RemotePeer, timeoutRPC time.Duration) Comm {
	commInst := &commImpl{
		server:          server,
		self:            self,
		subscriptions:   make([]*subscription, 0),
		lock:            sync.Mutex{},
		stopping:        int32(0),
		exitChan:        make(chan struct{}),
		deMuxInProgress: sync.WaitGroup{},
		sendInProgress:  sync.WaitGroup{},
		timeoutRPC:      timeoutRPC,
		logger:          log.GetLogger(self.Endpoint),
	}

	// register the message handler to rpcServer
	err := server.Register(&MessageHandler{
		comm: commInst,
	})
	if err != nil {
		commInst.logger.Error("register rpc service failed")
	}
	return commInst
}

type commImpl struct {
	server          *rpc.Server
	self            *proto.RemotePeer
	subscriptions   []*subscription
	lock            sync.Mutex
	stopping        int32
	exitChan        chan struct{}
	deMuxInProgress sync.WaitGroup
	sendInProgress  sync.WaitGroup
	timeoutRPC      time.Duration
	logger          *zap.Logger
}

func (c *commImpl) HandleMessage(message *proto.ReceivedMessage) {
	c.deMultiplex(message)
}

func (c *commImpl) deMultiplex(message *proto.ReceivedMessage) {
	c.lock.Lock()
	if c.isStopping() {
		c.lock.Unlock()
		return
	}
	subscriptions := c.subscriptions
	c.deMuxInProgress.Add(1)
	c.lock.Unlock()

	for _, sub := range subscriptions {
		if sub.pred(message) {
			select {
			case <-c.exitChan:
				c.deMuxInProgress.Done()
				return
			case sub.ch <- message:
			}
		}
	}
	c.deMuxInProgress.Done()
}

func (c *commImpl) Self() *proto.RemotePeer {
	return c.self
}

func (c *commImpl) Send(msg *proto.Envelope[*proto.Message], peers ...*proto.RemotePeer) {
	if c.isStopping() || len(peers) == 0 {
		return
	}
	c.logger.Debug(fmt.Sprintf("begin to send message to %d peers", len(peers)))

	c.sendInProgress.Add(len(peers))
	for _, peer := range peers {
		go func(peer *proto.RemotePeer, msg *proto.Envelope[*proto.Message]) {
			c.sendToEndpoint(peer, msg)
		}(peer, msg)
	}
}

func (c *commImpl) sendToEndpoint(peer *proto.RemotePeer, msg *proto.Envelope[*proto.Message]) {
	defer c.sendInProgress.Done()
	conn, err := rpc.Dial("tcp", peer.Endpoint)
	if err != nil {
		c.logger.Error("failed to dial peer " + peer.Endpoint + ", err = " + err.Error())
		return
	}
	var msgToSend = &proto.ReceivedMessage{
		Envelope: msg,
		Sender:   c.self,
	}
	var reply proto.ReceivedMessage

	defer func(conn *rpc.Client) {
		_ = conn.Close()
	}(conn)
	errChan := make(chan error)
	go func() {
		c.logger.Info("call MessageHandler.HandleMessage with " + peer.Endpoint)
		errChan <- conn.Call("MessageHandler.HandleMessage", msgToSend, &reply)
	}()

	select {
	case err = <-errChan:
		if err != nil {
			c.logger.Error("rpc failed, err: " + err.Error())
		}
	case <-time.After(c.timeoutRPC):
		c.logger.Error("connection to " + peer.Endpoint + " failed, timeout")
	}

}

func (c *commImpl) Accept(acceptor MessageAcceptor) <-chan *proto.ReceivedMessage {
	messageChan := make(chan *proto.ReceivedMessage, 10)

	if c.isStopping() {
		fmt.Printf("return an empty channel because of the comm is stopped")
		return messageChan
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	sub := &subscription{
		ch:   messageChan,
		pred: acceptor,
	}
	c.subscriptions = append(c.subscriptions, sub)
	return messageChan
}

func (c *commImpl) closeSubscriptions() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, sub := range c.subscriptions {
		close(sub.ch)
	}
}

func (c *commImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&c.stopping, 0, int32(1)) {
		return
	}
	fmt.Println("Stopping")
	defer fmt.Println("Stopped")
	close(c.exitChan)
	c.closeSubscriptions()
}

func (c *commImpl) isStopping() bool {
	return atomic.LoadInt32(&c.stopping) == int32(1)
}
