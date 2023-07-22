package python

import "github.com/wisecoach/ml_chain/proto"

type AggregateRequest struct {
	LocalModels []*proto.LocalityWeight `json:"local_models,omitempty"`
}

type AggregateResponse struct {
	GlobalModel *proto.GlobalWeight `json:"global_model,omitempty"`
}

type ValidateRequest struct {
	Model *proto.LocalityWeight
}

type ValidateResponse struct {
	Loss *proto.ValidateLoss
}

type TrainRequest struct {
	GlobalModel *proto.GlobalWeight
}

type TrainResponse struct {
	LocalModel *proto.LocalityWeight
}
