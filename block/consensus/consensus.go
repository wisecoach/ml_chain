package consensus

import "github.com/wisecoach/ml_chain/proto"

type Consensus interface {
	//
	// Order
	//  @Description: order a transaction
	//
	Order(transaction *proto.Envelope[*proto.Transaction])

	// Consensus
	//  @Description: consensus a block, validate if block is valid and confirm to blockMgr
	//
	Consensus(block *proto.Envelope[*proto.Block])

	//
	// Start
	//  @Description: start the consensus service
	//
	Start()
}
