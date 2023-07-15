package role

import (
	"github.com/wisecoach/ml_chain/comm/comm"
	"sync"
)

type Manager interface {
	//
	// SelectVerifiers
	//  @Description: select the verifiers
	//
	SelectVerifiers(candidates []*comm.RemotePeer, number int)

	//
	// GetVerifiers
	//  @Description: get the selected verifiers
	//  @return []*comm.RemotePeer
	//
	GetVerifiers() []*comm.RemotePeer

	//
	// AmVerifier
	//  @Description: if self is verifier
	//  @return bool
	//
	AmVerifier() bool
}

type roleMgr struct {
	lock sync.RWMutex

	verifiers []*comm.RemotePeer

	selector *Selector
}

func (r *roleMgr) SelectVerifiers(candidates []*comm.RemotePeer, input []byte, number int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.verifiers = r.selector.SelectVerifiers(candidates, input, number)
}

func (r *roleMgr) GetVerifiers() []*comm.RemotePeer {
	// TODO implement me
	panic("implement me")
}

func (r *roleMgr) AmVerifier() bool {
	// TODO implement me
	panic("implement me")
}
