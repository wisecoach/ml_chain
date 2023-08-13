package task

import (
	"encoding/base64"
	"github.com/wisecoach/ml_chain/proto"
	"sync"
)

type taskManager struct {
	lock sync.Mutex

	managers     map[string]struct{}
	tasks        map[string]*Task
	taskFinished map[string]*FinishedTask
}

func NewTaskManager() Manager {
	return &taskManager{
		lock:         sync.Mutex{},
		managers:     make(map[string]struct{}),
		tasks:        make(map[string]*Task),
		taskFinished: make(map[string]*FinishedTask),
	}
}

func (t *taskManager) CreateTask(task *proto.TaskGenesis) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.tasks[task.TaskId] = &Task{TaskGenesis: task}
}

func (t *taskManager) FinishTask(result *proto.TaskResult) {
	t.lock.Lock()
	defer t.lock.Unlock()

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

	key := base64.StdEncoding.EncodeToString(pk)
	t.managers[key] = struct{}{}
}

func (t *taskManager) RevokeManager(pk []byte) {
	t.lock.Lock()
	defer t.lock.Unlock()

	key := base64.StdEncoding.EncodeToString(pk)
	delete(t.managers, key)
}
