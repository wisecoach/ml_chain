package txhandler

import (
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/iteration"
	"reflect"
)

type IterationTxHandler struct {
	iterationManager iteration.Manager
}

func NewIterationTxHandler(iterationManager iteration.Manager) manager.TxHandler {
	return &IterationTxHandler{iterationManager: iterationManager}
}

func (i *IterationTxHandler) HandleTx(tx *proto.Transaction) {
	modelIteration := tx.Payload.(*proto.ModelIteration)
	i.iterationManager.NextIteration(modelIteration)
}

func (i *IterationTxHandler) TxType() reflect.Type {
	return reflect.TypeOf(&proto.ModelIteration{})
}
