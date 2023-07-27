package aggregate

import "github.com/wisecoach/ml_chain/proto"

type Aggregator interface {
	// HandleLocalModel
	//  @Description: handle the local model sent from trainer
	HandleLocalModel(weight *proto.LocalityWeight) error

	// StartAggregate
	//  @Description: start aggregate:
	//  				1. wait for local model to aggregate,
	//  				2. aggregate the new global model
	//  				3. send ModelIteration transaction to consensus module
	StartAggregate()
}
