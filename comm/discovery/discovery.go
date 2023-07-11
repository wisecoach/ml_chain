package discovery

import (
	"github.com/wisecoach/ml_chain/comm/comm"
)

// Discovery is the interface that represents a discovery module
type Discovery interface {
	// Lookup returns a network member, or nil if not found
	Lookup(pk []byte) *comm.RemotePeer

	// Register register a remote peer
	Register(peer *comm.RemotePeer)

	// Self returns this instance's membership information
	Self() *comm.RemotePeer

	// GetMembership returns the alive members in the view
	GetMembership() []*comm.RemotePeer
}

type discoveryImpl struct {
	me    *comm.RemotePeer
	peers []*comm.RemotePeer
}

func New(me *comm.RemotePeer) Discovery {
	return &discoveryImpl{
		me: me,
	}
}

func (d discoveryImpl) Lookup(pk []byte) *comm.RemotePeer {
	// TODO implement me
	panic("implement me")
}

func (d discoveryImpl) Register(peer *comm.RemotePeer) {
	d.peers = append(d.peers, peer)
}

func (d discoveryImpl) Self() *comm.RemotePeer {
	return d.me
}

func (d discoveryImpl) GetMembership() []*comm.RemotePeer {
	return d.peers
}
