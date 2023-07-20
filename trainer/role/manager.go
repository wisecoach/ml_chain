package role

import (
	set "github.com/deckarep/golang-set/v2"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/proto"
	"sync"
)

type Manager interface {
	//
	// SelectValidators
	//  @Description: select the validatorSelectionResults
	//
	SelectValidators(candidates []*comm.RemotePeer, number int)

	//
	// GetValidators
	//  @Description: get the selected validatorSelectionResults
	//  @return *SelectionResult
	//
	GetValidators() *proto.SelectionResult

	//
	// AmValidator
	//  @Description: if self is validatorSelectionResult
	//  @return bool
	//
	AmValidator() bool
}

type roleMgr struct {
	lock sync.RWMutex

	self *comm.RemotePeer

	validatorSelectionResult *proto.SelectionResult
	validatorSet             set.Set[*comm.RemotePeer]

	selector *Selector
}

func (r *roleMgr) SelectValidators(candidates []*comm.RemotePeer, input []byte, number int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.validatorSelectionResult = r.selector.SelectValidators(candidates, input, number)
	r.validatorSet = set.NewSet(r.validatorSelectionResult.Winners...)
}

func (r *roleMgr) GetValidators() *proto.SelectionResult {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.validatorSelectionResult
}

func (r *roleMgr) AmValidator() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.validatorSet.Contains(r.self)
}
