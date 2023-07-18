package consensus

import "github.com/wisecoach/ml_chain/proto"

type Consensus interface {
	//
	// Order
	//  @Description: order a transaction
	//
	Order(transaction *proto.Envelope[proto.Transaction])

	//
	// Start
	//  @Description: start the consensus service
	//
	Start()
}
