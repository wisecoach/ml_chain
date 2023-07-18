package manager

import (
	"github.com/wisecoach/ml_chain/block/chain"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
)

type BlockHandler interface {
	HandleBlock(block *proto.Block)
}

type TxHandler interface {
	HandleTx(tx *proto.Transaction)
	// TxType
	//  @Description: get the type which
	TxType() reflect.Type
}

type BlockValidator interface {
	ValidateBlock(block *proto.Block) error
}

type TxValidator interface {
	ValidateTx(tx *proto.Envelope[*proto.Transaction]) error
}

type BlockManager interface {
	//
	// ConfirmBlock
	//  @Description: confirm a block, it means the block has been confirmed by consensus and will be invoked by consensus module
	// 				  confirm a block will add block to the blockchain and signal all the blockhandler and txhandler register in manager
	ConfirmBlock(block *proto.Block) error
	GetBlock(number int) *proto.Block

	//
	// CreateBlock
	//  @Description: create a block with the txs, it should be invoked by consensus module which will collect tx from other peer
	// 					it's will be removed because of the blockchain is not contained the tip block but the confirmed block
	//
	// CreateBlock(txs []*data.Transaction) *data.Block
	GetChain() *chain.BlockChain
	RegisterBlockHandler(handler BlockHandler)
	RegisterTxHandler(handler TxHandler)
}
