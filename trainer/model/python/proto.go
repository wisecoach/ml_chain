package python

import "github.com/wisecoach/ml_chain/proto"

type AggregateRequest struct {
	LocalModels []*proto.LocalityWeight `json:"local_models,omitempty"`
}

type AggregateResponse struct {
	GlobalModel *proto.GlobalWeight `json:"global_model,omitempty"`
}

type ValidateRequest struct {
	Model *proto.LocalityWeight `json:"model"`
}

type ValidateResponse struct {
	Loss *proto.ValidateLoss `json:"loss"`
}

type TrainRequest struct {
	TrainerId   string              `json:"trainer_id"`
	GlobalModel *proto.GlobalWeight `json:"global_model"`
}

type TrainResponse struct {
	LocalModel *proto.LocalityWeight `json:"local_model"`
}
