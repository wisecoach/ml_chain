package discovery

import (
	"bytes"
	"github.com/wisecoach/ml_chain/proto"
	"sync"
)

// Discovery is the interface that represents a discovery module
type Discovery interface {
	// Lookup returns a network member, or nil if not found
	Lookup(pk []byte) *proto.RemotePeer

	// Register register a remote peer
	Register(peer *proto.RemotePeer)

	// Self returns this instance's membership information
	Self() *proto.RemotePeer

	// GetMembership returns the alive members in the view
	GetMembership() []*proto.RemotePeer
}

type discoveryImpl struct {
	lock  *sync.RWMutex
	self  *proto.RemotePeer
	peers map[string]*proto.RemotePeer
}

func New(self *proto.RemotePeer) Discovery {
	return &discoveryImpl{
		lock:  &sync.RWMutex{},
		self:  self,
		peers: make(map[string]*proto.RemotePeer),
	}
}

func (d discoveryImpl) Lookup(pk []byte) *proto.RemotePeer {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.peers[string(pk)]
}

func (d discoveryImpl) Register(peer *proto.RemotePeer) {
	if bytes.Equal(peer.PublicKey, d.self.PublicKey) {
		return
	}
	d.lock.Lock()
	defer d.lock.Unlock()

	d.peers[string(peer.PublicKey)] = peer
}

func (d discoveryImpl) Self() *proto.RemotePeer {
	return d.self
}

func (d discoveryImpl) GetMembership() []*proto.RemotePeer {
	d.lock.RLock()
	defer d.lock.RUnlock()

	membership := make([]*proto.RemotePeer, len(d.peers))
	for _, peer := range d.peers {
		membership = append(membership, peer)
	}
	return membership
}
