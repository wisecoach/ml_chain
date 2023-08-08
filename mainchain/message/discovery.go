package message

import (
	"github.com/wisecoach/ml_chain/comm/discovery"
	"github.com/wisecoach/ml_chain/proto"
)

type PeerRegisterListener struct {
	disc discovery.Discovery
}

func NewPeerRegisterListener(disc discovery.Discovery) *PeerRegisterListener {
	return &PeerRegisterListener{
		disc: disc,
	}
}

func (t *PeerRegisterListener) HandleMessage(message *proto.ReceivedMessage) {
	peerRegister := message.Envelope.Payload.GetPeerRegitser()
	t.disc.Register(peerRegister.Peer)
}
