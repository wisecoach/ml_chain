package role

import (
	set "github.com/deckarep/golang-set/v2"
	"github.com/wisecoach/ml_chain/comm/comm"
	"sync"
)

type Manager interface {
	//
	// SelectVerifiers
	//  @Description: select the verifierSelectionResults
	//
	SelectVerifiers(candidates []*comm.RemotePeer, number int)

	//
	// GetVerifiers
	//  @Description: get the selected verifierSelectionResults
	//  @return *SelectionResult
	//
	GetVerifiers() *SelectionResult

	//
	// AmVerifier
	//  @Description: if self is verifierSelectionResult
	//  @return bool
	//
	AmVerifier() bool
}

type roleMgr struct {
	lock sync.RWMutex

	self *comm.RemotePeer

	verifierSelectionResult *SelectionResult
	verifierSet             set.Set[*comm.RemotePeer]

	selector *Selector
}

func (r *roleMgr) SelectVerifiers(candidates []*comm.RemotePeer, input []byte, number int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.verifierSelectionResult = r.selector.SelectVerifiers(candidates, input, number)
	r.verifierSet = set.NewSet(r.verifierSelectionResult.winners...)
}

func (r *roleMgr) GetVerifiers() *SelectionResult {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.verifierSelectionResult
}

func (r *roleMgr) AmVerifier() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.verifierSet.Contains(r.self)
}
