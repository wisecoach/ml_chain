package proto

type GlobalWeight struct {
	Iteration    int       `json:"iteration"`
	WeightVector []float64 `json:"weight_vector"`
	Aggregator   []byte    `json:"aggregator"`
}

type LocalityWeight struct {
	Iteration          int                        `json:"iteration"`
	WeightVector       []float64                  `json:"weight_vector"`
	Trainer            []byte                     `json:"trainer"`
	ValidatorSelection *SelectionResult           `json:"validator_selection"`
	Losses             []*Envelope[*ValidateLoss] `json:"losses"`
	Acc                float64                    `json:"acc"`
	Loss               float64                    `json:"loss"`
	DataNum            int                        `json:"data_num"`
}

type ValidateLoss struct {
	Iteration int     `json:"iteration"`
	Trainer   []byte  `json:"trainer"`
	Validator []byte  `json:"validator"`
	ModelHash []byte  `json:"model_hash"`
	Loss      float64 `json:"loss"`
	Acc       float64 `json:"acc"`
}
