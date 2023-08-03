package role

import (
	set "github.com/deckarep/golang-set/v2"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
	"sync"
)

type Manager interface {
	//
	// AmAggregator
	//  @Description: check if self is the aggregator for the iteration
	//
	AmAggregator(iteration int) bool

	// GetAggregator
	//  @Description: get the aggregator for the iteration
	//
	GetAggregator(iteration int) []byte

	// SelectValidators
	//  @Description: select the validatorSelectionResults
	//
	SelectValidators(candidates []*proto.RemotePeer, input []byte, number int)

	// GetValidators
	//  @Description: get the selected validatorSelectionResults
	//  @return *SelectionResult
	//
	GetValidators() *proto.SelectionResult

	// AmValidator
	//  @Description: if self is validatorSelectionResult
	//  @return bool
	//
	AmValidator(selection *proto.SelectionResult) bool

	// VerifyValidatorSelection
	//  @Description: verify the validator selection
	//
	VerifyValidatorSelection(proof *proto.SelectionResult) error
}

func New(config *Config) Manager {
	r := &roleMgr{
		lock:            sync.RWMutex{},
		config:          config,
		self:            config.Self,
		taskManagerList: config.TaskManagerList,
		selector:        &Selector{},
	}
	r.selector.init()
	return r
}

type roleMgr struct {
	lock sync.RWMutex

	self            *proto.RemotePeer
	taskManagerList [][]byte
	config          *Config

	validatorSelectionResult *proto.SelectionResult
	validatorSet             set.Set[*proto.RemotePeer]

	selector *Selector
}

func (r *roleMgr) AmAggregator(iteration int) bool {
	return reflect.DeepEqual(r.self.PublicKey, r.GetAggregator(iteration))
}

func (r *roleMgr) GetAggregator(iteration int) []byte {
	num := len(r.taskManagerList)
	return r.taskManagerList[iteration%num]
}

func (r *roleMgr) SelectValidators(candidates []*proto.RemotePeer, input []byte, number int) {
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

func (r *roleMgr) AmValidator(selection *proto.SelectionResult) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	err := r.selector.VerifyValidatorSelection(selection)
	if err != nil {
		return false
	}
	return set.NewSet(r.validatorSelectionResult.Winners...).Contains(r.self)
}

func (r *roleMgr) VerifyValidatorSelection(proof *proto.SelectionResult) error {
	return r.selector.VerifyValidatorSelection(proof)
}
