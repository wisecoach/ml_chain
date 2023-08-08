package message

import (
	"github.com/wisecoach/ml_chain/comm/discovery"
	"github.com/wisecoach/ml_chain/proto"
)

type PeerRegitserListener struct {
	disc discovery.Discovery
}

func NewPeerRegitserListener(disc discovery.Discovery) *PeerRegitserListener {
	return &PeerRegitserListener{
		disc: disc,
	}
}

func (t *PeerRegitserListener) HandleMessage(message *proto.ReceivedMessage) {
	PeerRegitser := message.Envelope.Payload.GetPeerRegitser()
	t.disc.Register(PeerRegitser.Peer)
}
