package message

import (
	"github.com/wisecoach/ml_chain/comm/discovery"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"time"
)

type PeerRegisterListener struct {
	disc discovery.Discovery
	node *node.Node
}

func NewPeerRegisterListener(disc discovery.Discovery, node *node.Node) *PeerRegisterListener {
	return &PeerRegisterListener{
		disc: disc,
		node: node,
	}
}

func (t *PeerRegisterListener) HandleMessage(message *proto.ReceivedMessage) {
	peerRegister := message.Envelope.Payload.GetPeerRegitser()
	if t.disc.Lookup(peerRegister.Peer.PublicKey) == nil {
		t.disc.Register(peerRegister.Peer)
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
