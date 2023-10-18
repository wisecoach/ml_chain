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
	TaskId     string                   `json:"task_id"`
	InitWeight *Envelope[*GlobalWeight] `json:"init_weight"`
}

// TaskResult represent the finish transaction of a task
type TaskResult struct {
	TaskId string
}

// ModelIteration represent a iteration of global model
type ModelIteration struct {
	Iteration       int
	GlobalWeight    *GlobalWeight
	LocalityWeights map[string]*LocalityWeight
	Contributes     map[string]float32
}

type ModelStructure struct {
	Dataset      string  `json:"dataset"`
	NumClasses   int     `json:"num_classes"`
	Agent        int     `json:"agent"`
	TrainerNum   int     `json:"trainer_num"`
	ValidatorNum int     `json:"validator_num"`
	LearningRate float32 `json:"learning_rate"`
	Momentum     float32 `json:"momentum"`
	Dp           bool    `json:"dp"`
	DpEpsilon    float32 `json:"dp_epsilon"`
	DpEpsilon1   float32 `json:"dp_epsilon_1"`
	DpDelta      float32 `json:"dp_delta"`
	DpClip       float32 `json:"dp_clip"`
	BatchSize    int     `json:"batch_size"`
	Round        int     `json:"round"`
	Lambda       float32 `json:"lambda"`
}
