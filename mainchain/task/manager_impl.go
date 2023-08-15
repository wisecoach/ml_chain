package task

import (
	"encoding/base64"
	"fmt"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"sync"
)

type taskManager struct {
	lock   sync.RWMutex
	logger *zap.Logger

	self         *proto.RemotePeer
	managers     map[string]struct{}
	tasks        map[string]*Task
	taskFinished map[string]*FinishedTask
}

func NewTaskManager(self *proto.RemotePeer) Manager {
	return &taskManager{
		lock:         sync.RWMutex{},
		logger:       log.GetLogger(self.Endpoint),
		self:         self,
		managers:     make(map[string]struct{}),
		tasks:        make(map[string]*Task),
		taskFinished: make(map[string]*FinishedTask),
	}
}

func (t *taskManager) GetManagers() [][]byte {
	t.lock.RLock()
	defer t.lock.RUnlock()

	managers := make([][]byte, len(t.managers))
	i := 0
	for manager, _ := range t.managers {
		managers[i], _ = base64.StdEncoding.DecodeString(manager)
		i += 1
	}
	return managers
}

func (t *taskManager) CreateTask(task *proto.TaskGenesis) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.logger.Info(fmt.Sprintf("create task %s", task.TaskId))
	t.tasks[task.TaskId] = &Task{TaskGenesis: task}
}

func (t *taskManager) FinishTask(result *proto.TaskResult) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.logger.Info(fmt.Sprintf("finish task %s", result.TaskId))
	task := t.tasks[result.TaskId]
	delete(t.tasks, result.TaskId)
	t.taskFinished[task.TaskGenesis.TaskId] = &FinishedTask{
		TaskGenesis: task.TaskGenesis,
		TaskResult:  result,
	}
}

func (t *taskManager) RegisterManager(pk []byte) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.logger.Info(fmt.Sprintf("register manager: %s", base64.StdEncoding.EncodeToString(pk)))
	key := base64.StdEncoding.EncodeToString(pk)
	t.managers[key] = struct{}{}
}

func (t *taskManager) RevokeManager(pk []byte) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.logger.Info(fmt.Sprintf("revoke manager: %s", base64.StdEncoding.EncodeToString(pk)))
	key := base64.StdEncoding.EncodeToString(pk)
	delete(t.managers, key)
}
