package task

type Client interface { //
	// CreateTask
	//  @Description: create a task, send a transaction to consensus module
	//  @param task
	//
	CreateTask(task *Task)

	//
	// FinishTask
	//  @Description: finish a task, send a transaction to consensus module
	//  @param result
	//
	FinishTask(task *FinishedTask)

	//
	// RegisterManager
	//  @Description: register manager, send a transaction to consensus module
	//  @param peer
	//
	RegisterManager(pk []byte)

	//
	// RevokeManager
	//  @Description: revoke manager, send a transaction to consensus module
	//  @param peer
	//
	RevokeManager(pk []byte)
}
