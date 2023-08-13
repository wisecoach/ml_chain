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

	// GetHeight
	//  @Description: get the height of block
	GetHeight() int

	// ConfirmBlock
	//  @Description: confirm a block, it means the block has been confirmed by consensus and will be invoked by consensus module
	// 				  confirm a block will add block to the blockchain and signal all the blockhandler and txhandler register in manager
	ConfirmBlock(block *proto.Block) error

	// GetBlock
	//  @Description: get the block which BlockNumber is number
	//
	GetBlock(number int) (*proto.Block, error)

	// GetLatestBlock
	//  @Description:
	//  @return *proto.Block
	//
	GetLatestBlock() (*proto.Block, error)

	// GetChain
	//  @Description: get the blockchain
	//
	GetChain() *chain.BlockChain

	// RegisterBlockHandler
	//  @Description: register the block handler
	//
	RegisterBlockHandler(handler BlockHandler)

	// RegisterTxHandler
	//  @Description:
	//
	RegisterTxHandler(handler TxHandler)
}
