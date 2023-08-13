package txhandler

import (
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/mainchain/task"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
)

func NewManagerRegisterTxHandler(taskManager task.Manager) manager.TxHandler {
	return &ManagerRegisterTxHandler{
		taskManager: taskManager,
	}
}

func NewManagerRevokeTxHandler(taskManager task.Manager) manager.TxHandler {
	return &ManagerRevokeTxHandler{
		taskManager: taskManager,
	}
}

type ManagerRegisterTxHandler struct {
	taskManager task.Manager
}

func (m *ManagerRegisterTxHandler) HandleTx(tx *proto.Transaction) {
	register := tx.Payload.(*proto.ManagerRegister)
	m.taskManager.RegisterManager(register.Registrar)
}

func (m *ManagerRegisterTxHandler) TxType() reflect.Type {
	return reflect.TypeOf(&proto.ManagerRegister{})
}

type ManagerRevokeTxHandler struct {
	taskManager task.Manager
}

func (m *ManagerRevokeTxHandler) HandleTx(tx *proto.Transaction) {
	revoke := tx.Payload.(*proto.ManagerRevoke)
	m.taskManager.RevokeManager(revoke.Manager)
}

func (m *ManagerRevokeTxHandler) TxType() reflect.Type {
	return reflect.TypeOf(&proto.ManagerRevoke{})
}
