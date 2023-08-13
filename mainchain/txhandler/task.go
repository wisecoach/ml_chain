package txhandler

import (
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/mainchain/task"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
)

func NewTaskGenesisTxHandler(taskManager task.Manager) manager.TxHandler {
	return &TaskGenesisTxHandler{
		taskManager: taskManager,
	}
}

func NewTaskFinishTxHandler(taskManager task.Manager) manager.TxHandler {
	return &TaskFinishTxHandler{
		taskManager: taskManager,
	}
}

type TaskGenesisTxHandler struct {
	taskManager task.Manager
}

func (t *TaskGenesisTxHandler) HandleTx(tx *proto.Transaction) {
	taskGenesis := tx.Payload.(*proto.TaskGenesis)
	t.taskManager.CreateTask(taskGenesis)
}

func (t *TaskGenesisTxHandler) TxType() reflect.Type {
	return reflect.TypeOf(&proto.TaskGenesis{})
}

type TaskFinishTxHandler struct {
	taskManager task.Manager
}

func (t *TaskFinishTxHandler) HandleTx(tx *proto.Transaction) {
	taskResult := tx.Payload.(*proto.TaskResult)
	t.taskManager.FinishTask(taskResult)
}

func (t *TaskFinishTxHandler) TxType() reflect.Type {
	return reflect.TypeOf(&proto.TaskResult{})
}
