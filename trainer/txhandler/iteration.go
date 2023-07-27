package txhandler

import (
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/iteration"
	"reflect"
	"sync"
)

type IterationTxHandler struct {
	lock             sync.Locker
	iterationManager iteration.Manager
}

func (i *IterationTxHandler) HandleTx(tx *proto.Transaction) {
	i.lock.Lock()
	defer i.lock.Unlock()

	modelIteration := tx.Payload.(*proto.ModelIteration)
	i.iterationManager.NextIteration(modelIteration)
}

func (i IterationTxHandler) TxType() reflect.Type {
	return reflect.TypeOf(&proto.ModelIteration{})
}
