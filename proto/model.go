package proto

type GlobalWeight struct {
	Iteration    int       `json:"iteration"`
	WeightVector []float32 `json:"weight_vector"`
	Aggregator   []byte    `json:"aggregator"`
	TotalAcc     float32   `json:"total_acc"`
	Loss         float32   `json:"loss"`
}

type LocalityWeight struct {
	Iteration          int                        `json:"iteration"`
	WeightVector       []float32                  `json:"weight_vector"`
	Trainer            []byte                     `json:"trainer"`
	ValidatorSelection *SelectionResult           `json:"validator_selection"`
	Losses             []*Envelope[*ValidateLoss] `json:"losses"`
	Acc                float32                    `json:"acc"`
	Loss               float32                    `json:"loss"`
	DataNum            int                        `json:"data_num"`
}

type ValidateLoss struct {
	Iteration int     `json:"iteration"`
	Trainer   []byte  `json:"trainer"`
	Validator []byte  `json:"validator"`
	ModelHash []byte  `json:"model_hash"`
	Loss      float32 `json:"loss"`
	Acc       float32 `json:"acc"`
}
