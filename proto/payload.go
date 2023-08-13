package proto

// Payload
// @Description: represent the payload of transaction
type Payload interface{}

type ManagerRegister struct {
	Registrar []byte
}

type ManagerRevoke struct {
	Manager []byte
}

// TaskGenesis represent the genesis transaction of a taskchain
type TaskGenesis struct {
	TaskId         string
	ModelStructure *ModelStructure
	ManagerList    [][]byte
	InitWeight     *Envelope[*GlobalWeight]
}

// TaskFinish represent the finish transaction of a task
type TaskResult struct {
	TaskId string
}

// ModelIteration represent a iteration of global model
type ModelIteration struct {
	Iteration       int
	GlobalWeight    *GlobalWeight
	LocalityWeights []*LocalityWeight
}

type ModelStructure struct {
	// TODO
}
