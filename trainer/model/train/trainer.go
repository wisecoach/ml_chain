package train

import "github.com/wisecoach/ml_chain/proto"

type LocalTrainer interface {
	// Train
	//  @Description: train the global model and begin to send the model to validator selected by vrf for validating,
	//				  with the prove of validating, send the local model to aggregator committee
	Train(weight *proto.GlobalWeight)
}
