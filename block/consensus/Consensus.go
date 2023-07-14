package consensus

import "github.com/wisecoach/ml_chain/block/data"

type Consensus interface {
	//
	// Order
	//  @Description: order a transaction
	//
	Order(transaction *data.SignedTransaction)

	//
	// Start
	//  @Description: start the consensus service
	//
	Start()
}
