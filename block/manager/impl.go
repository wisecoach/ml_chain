package manager

import (
	"github.com/wisecoach/ml_chain/block/data"
	"reflect"
	"sync"
)

type blockMgr struct {
	bc            *data.BlockChain
	blockHandlers []BlockHandler
	txHandlers    map[reflect.Type][]*TxHandler

	lock sync.RWMutex
}

func (b blockMgr) ConfirmBlock(block *data.Block) {
	// TODO implement me
	panic("implement me")
}

func (b blockMgr) GetBlock(number int) *data.Block {
	// TODO implement me
	panic("implement me")
}

func (b blockMgr) CreateBlock(txs []*data.Transaction) *data.Block {
	// TODO implement me
	panic("implement me")
}

func (b blockMgr) GetChain() *data.BlockChain {
	// TODO implement me
	panic("implement me")
}

func (b blockMgr) RegisterBlockHandler(handler *BlockHandler) {
	// TODO implement me
	panic("implement me")
}

func (b blockMgr) RegisterTxHandler(handler *TxHandler) {
	// TODO implement me
	panic("implement me")
}
