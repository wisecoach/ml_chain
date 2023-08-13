package task

import "github.com/wisecoach/ml_chain/proto"

type Task struct {
	TaskGenesis *proto.TaskGenesis
}

type FinishedTask struct {
	TaskGenesis *proto.TaskGenesis
	TaskResult  *proto.TaskResult
}
