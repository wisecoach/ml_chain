package comm

import (
	"fmt"
	"github.com/wisecoach/ml_chain/util/logger"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"
)

type MessageHandler struct {
	comm Comm
}

type subscription struct {
	ch   chan *ReceivedMessage
	pred MessageAcceptor
}

// HandleMessage
//
//	@Description:
//	@receiver mh
//	@param message
//	@param reply	it's always be nil, and shouldn't be operated
func (mh *MessageHandler) HandleMessage(message *ReceivedMessage, reply *ReceivedMessage) {
	mh.comm.HandleMessage(message)
}

func New(server *rpc.Server, self *RemotePeer, timeoutRPC time.Duration) Comm {
	commInst := &commImpl{
		server:     server,
		self:       self,
		exitChan:   make(chan struct{}),
		stopping:   int32(0),
		timeoutRPC: timeoutRPC,
	}

	// register the message handler to rpcServer
	err := server.Register(&MessageHandler{
		comm: commInst,
	})
	if err != nil {
		logger.Error("register rpc service failed")
	}
	return commInst
}

type commImpl struct {
	server          *rpc.Server
	self            *RemotePeer
	subscriptions   []*subscription
	lock            *sync.Mutex
	stopping        int32
	exitChan        chan struct{}
	deMuxInProgress sync.WaitGroup
	sendInProgress  sync.WaitGroup
	timeoutRPC      time.Duration
}

func (c *commImpl) HandleMessage(message *ReceivedMessage) {
	c.deMultiplex(message)
}

func (c *commImpl) deMultiplex(message *ReceivedMessage) {
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

func (c *commImpl) Self() *RemotePeer {
	return c.self
}

func (c *commImpl) Send(msg *SignedMessage, peers ...*RemotePeer) {
	if c.isStopping() || len(peers) == 0 {
		return
	}
	logger.Debug(fmt.Sprintf("begin to send message to %d peers", len(peers)))

	c.sendInProgress.Add(len(peers))
	for _, peer := range peers {
		go func(peer *RemotePeer, msg *SignedMessage) {
			c.sendToEndpoint(peer, msg)
		}(peer, msg)
	}
}

func (c *commImpl) sendToEndpoint(peer *RemotePeer, msg *SignedMessage) {
	defer c.sendInProgress.Done()
	conn, err := rpc.Dial("tcp", peer.Endpoint)
	if err != nil {
		logger.Error("failed to dial peer " + peer.Endpoint + ", err = " + err.Error())
		return
	}
	var msgToSend = &ReceivedMessage{
		SignedMessage: msg,
		sender:        c.self,
	}
	var reply ReceivedMessage

	defer func(conn *rpc.Client) {
		_ = conn.Close()
	}(conn)
	errChan := make(chan error)
	go func() {
		errChan <- conn.Call("MessageHandler.HandleMessage", msgToSend, &reply)
	}()

	select {
	case err = <-errChan:
		if err != nil {
			logger.Error("rpc failed, err: " + err.Error())
		}
	case <-time.After(c.timeoutRPC):
		logger.Error("connection to " + peer.Endpoint + " failed, timeout")
	}

}

func (c *commImpl) Accept(acceptor MessageAcceptor) <-chan *ReceivedMessage {
	messageChan := make(chan *ReceivedMessage, 10)

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
