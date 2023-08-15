package task

import "github.com/wisecoach/ml_chain/proto"

type Manager interface {
	//
	// GetManagers
	//  @Description: get all the managers
	//
	GetManagers() [][]byte
	//
	// CreateTask
	//  @Description: create a task, when handle the TaskGenesisTransaction
	//  @param task
	//
	CreateTask(task *proto.TaskGenesis)

	//
	// FinishTask
	//  @Description: finish a task, when handle the TaskResultTransaction,
	//  @param result
	//
	FinishTask(task *proto.TaskResult)

	//
	// RegisterManager
	//  @Description: register manager, when handle the RegisterManagerTransaction
	//  @param peer
	//
	RegisterManager(pk []byte)

	//
	// RevokeManager
	//  @Description: revoke manager, when handle the UnRegisterManagerTransaction
	//  @param peer
	//
	RevokeManager(pk []byte)
}
