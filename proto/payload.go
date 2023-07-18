package proto

// Payload
// @Description: represent the payload of transaction
type Payload struct {
}

// TaskGenesis 任务的创世区块
type TaskGenesis struct {
	Payload
}

// ModelIteration 一轮全局模型的迭代
type ModelIteration struct {
	Payload
}
