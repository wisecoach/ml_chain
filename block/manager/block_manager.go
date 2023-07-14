package manager

import (
	"github.com/wisecoach/ml_chain/block/data"
	"reflect"
)

type BlockHandler interface {
	HandleBlock(block *data.Block)
}

type TxHandler interface {
	HandleTx(tx *data.Transaction)
	// TxType
	//  @Description:
	TxType() reflect.Type
}

type BlockManager interface {
	//
	// ConfirmBlock
	//  @Description: confirm a block, it means the block has been confirmed by consensus and will be invoked by consensus module
	// 				  confirm a block will add block to the blockchain and signal all the blockhandler and txhandler register in manager
	ConfirmBlock(block *data.Block)
	GetBlock(number int) *data.Block

	//
	// CreateBlock
	//  @Description: create a block with the txs, it should be invoked by consensus module which will collect tx from other peer
	//
	CreateBlock(txs []*data.Transaction) *data.Block
	GetChain() *data.BlockChain
	RegisterBlockHandler(handler *BlockHandler)
	RegisterTxHandler(handler *TxHandler)
}
