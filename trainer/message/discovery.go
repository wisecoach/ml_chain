package message

import (
	"fmt"
	"github.com/wisecoach/ml_chain/comm/discovery"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"sync"
	"time"
)

type PeerRegisterListener struct {
	disc      discovery.Discovery
	node      *node.Node
	logger    *zap.Logger
	waitGroup sync.WaitGroup
}

func NewPeerRegisterListener(numToWait int, disc discovery.Discovery, node *node.Node) *PeerRegisterListener {
	listener := &PeerRegisterListener{
		disc:   disc,
		node:   node,
		logger: log.GetLogger(node.Self().Endpoint),
	}
	listener.waitGroup.Add(numToWait)
	return listener
}

func (t *PeerRegisterListener) HandleMessage(message *proto.ReceivedMessage) {
	peerRegister := message.Envelope.Payload.GetPeerRegitser()
	if t.disc.Lookup(peerRegister.Peer.PublicKey) == nil {
		t.logger.Info(fmt.Sprintf("register peer: %s", peerRegister.Peer.Endpoint))
		t.disc.Register(peerRegister.Peer)
		t.waitGroup.Done()
		msg := &proto.Message{
			Content: &proto.PeerRegisterMessage{Peer: t.node.Self()},
			Header: &proto.Header{
				Creator:   t.node.Self().PublicKey,
				ChainId:   message.Envelope.Payload.Header.ChainId,
				TxId:      "",
				Timestamp: time.Time{},
			},
		}
		t.node.SendToPeers(msg, message.Sender)
	}
}

func (t *PeerRegisterListener) WaitForDiscover() {
	t.waitGroup.Wait()
}
