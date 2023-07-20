package proto

// Payload
// @Description: represent the payload of transaction
type Payload struct {
}

// TaskGenesis represent the genesis transaction of a taskchain
type TaskGenesis struct {
	Payload

	ModelStructure *ModelStructure
	ManagerList    [][]byte
	InitWeight     *Envelope[*GlobalWeight]
}

// ModelIteration represent a iteration of global model
type ModelIteration struct {
	Payload

	Iteration       int
	Accuracy        float64 // need?
	GlobalWeight    *Envelope[*GlobalWeight]
	LocalityWeights []*Envelope[*LocalityWeight]
}

type ModelStructure struct {
	// TODO
}
