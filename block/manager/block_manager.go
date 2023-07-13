package manager

import "github.com/wisecoach/ml_chain/block/data"

type BlockManager interface {
	AddBlock(newBlock *data.Block)
	GetBlock(number int) *data.Block
	CreateBlock(txs []*data.Transaction) *data.Block
	GetChain() *data.BlockChain
	RegisterBlockHandler
}
