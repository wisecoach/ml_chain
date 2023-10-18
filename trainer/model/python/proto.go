package python

import "github.com/wisecoach/ml_chain/proto"

type AggregateRequest struct {
	LocalModels     map[string]*proto.LocalityWeight `json:"local_models,omitempty"`
	LastGlobalModel *proto.GlobalWeight              `json:"last_global_model"`
}

type AggregateResponse struct {
	GlobalModel *proto.GlobalWeight `json:"global_model,omitempty"`
	Contributes map[string]float32  `json:"contributes"`
}

type ValidateRequest struct {
	Model *proto.LocalityWeight `json:"model"`
}

type ValidateResponse struct {
	Loss *proto.ValidateLoss `json:"loss"`
}

type TrainRequest struct {
	Cindex      string              `json:"cindex"`
	GlobalModel *proto.GlobalWeight `json:"global_model"`
}

type TrainResponse struct {
	LocalModel *proto.LocalityWeight `json:"local_model"`
}
