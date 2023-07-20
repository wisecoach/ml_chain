package proto

type GlobalWeight struct {
	Iteration    int
	WeightVector []float64
	Aggregator   []byte
}

type LocalityWeight struct {
	Iteration          int
	WeightVector       []float64
	Trainer            []byte
	ValidatorSelection *SelectionResult
	Losses             []*Envelope[*ValidateLoss]
}

type ValidateLoss struct {
	Validator []byte
	Loss      []float64 // loss vector
}
