package python

import "github.com/wisecoach/ml_chain/proto"

type Client interface {
	Init(genesis *proto.TaskGenesis) error
	//
	// Aggregate
	//  @Description: call python to aggregate the local models in request, and get the global model in response
	//
	Aggregate(request *AggregateRequest) (*AggregateResponse, error)
	//
	// Validate
	//  @Description: call python to validate the local model in request, and get the loss in response
	//
	Validate(request *ValidateRequest) (*ValidateResponse, error)
	//
	// Train
	//  @Description: call python to train the global model in request, and get the local model in response
	//
	Train(request *TrainRequest) (*TrainResponse, error)
}
