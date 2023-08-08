package txhandler

import (
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
)

func NewTaskGenesisTxHandler() manager.TxHandler {
	return &TaskGenesisTxHandler{}
}

type TaskGenesisTxHandler struct {
}

func (t TaskGenesisTxHandler) HandleTx(tx *proto.Transaction) {
	// TODO implement me
	panic("implement me")
}

func (t TaskGenesisTxHandler) TxType() reflect.Type {
	// TODO implement me
	panic("implement me")
}
