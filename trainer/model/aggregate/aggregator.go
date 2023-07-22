package aggregate

import "github.com/wisecoach/ml_chain/proto"

type Aggregator interface {
	//
	// HandleLocalModel
	//  @Description:
	//  @param weight
	//  @return error
	//
	HandleLocalModel(weight *proto.LocalityWeight) error
}
