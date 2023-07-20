package discovery

import (
	"bytes"
	"github.com/wisecoach/ml_chain/comm/comm"
	"sync"
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
	lock  *sync.RWMutex
	self  *comm.RemotePeer
	peers map[string]*comm.RemotePeer
}

func New(self *comm.RemotePeer) Discovery {
	return &discoveryImpl{
		lock:  &sync.RWMutex{},
		self:  self,
		peers: make(map[string]*comm.RemotePeer),
	}
}

func (d discoveryImpl) Lookup(pk []byte) *comm.RemotePeer {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.peers[string(pk)]
}

func (d discoveryImpl) Register(peer *comm.RemotePeer) {
	if bytes.Equal(peer.PublicKey, d.self.PublicKey) {
		return
	}
	d.lock.Lock()
	defer d.lock.Unlock()

	d.peers[string(peer.PublicKey)] = peer
}

func (d discoveryImpl) Self() *comm.RemotePeer {
	return d.self
}

func (d discoveryImpl) GetMembership() []*comm.RemotePeer {
	d.lock.RLock()
	defer d.lock.RUnlock()

	membership := make([]*comm.RemotePeer, len(d.peers))
	for _, peer := range d.peers {
		membership = append(membership, peer)
	}
	return membership
}
