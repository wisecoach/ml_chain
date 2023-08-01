package txhandler

import (
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/iteration"
	"reflect"
)

type GenesisTxHandler struct {
	iterationManager iteration.Manager
}

func NewGenesisTxHandler(iterationManager iteration.Manager) manager.TxHandler {
	return &GenesisTxHandler{iterationManager: iterationManager}
}

func (g *GenesisTxHandler) HandleTx(tx *proto.Transaction) {
	genesis := tx.Payload.(*proto.TaskGenesis)
	g.iterationManager.Start(genesis)
}

func (g *GenesisTxHandler) TxType() reflect.Type {
	return reflect.TypeOf(&proto.TaskGenesis{})
}
