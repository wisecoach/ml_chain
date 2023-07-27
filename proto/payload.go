package proto

// Payload
// @Description: represent the payload of transaction
type Payload interface{}

// TaskGenesis represent the genesis transaction of a taskchain
type TaskGenesis struct {
	ModelStructure *ModelStructure
	ManagerList    [][]byte
	InitWeight     *Envelope[*GlobalWeight]
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
