package manager

import (
	"github.com/wisecoach/ml_chain/block/chain"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/logger"
	"reflect"
	"sync"
)

type blockMgr struct {
	bc             *chain.BlockChain
	blockValidator BlockValidator
	txValidator    TxValidator
	blockHandlers  []BlockHandler
	txHandlers     map[reflect.Type][]TxHandler

	handlerLock sync.RWMutex
	bcLock      sync.RWMutex
}

func (b *blockMgr) ConfirmBlock(block *proto.Block) error {
	// validate the block
	err := b.blockValidator.ValidateBlock(block)
	if err != nil {
		return err
	}

	// handle the block and txs in the block
	b.handlerLock.RLock()
	blockHandlers := b.blockHandlers
	txHandlers := b.txHandlers
	b.handlerLock.RUnlock()

	for _, blockHandler := range blockHandlers {
		blockHandler.HandleBlock(block)
	}

	for _, signedTransaction := range block.Data.Transactions {
		tx := signedTransaction.Payload
		txType := reflect.TypeOf(tx.Payload)
		specHandlers, exists := txHandlers[txType]
		if exists {
			for _, txHandler := range specHandlers {
				txHandler.HandleTx(tx)
			}
		}
	}

	b.bcLock.Lock()
	// add block to blockchain
	b.bc.AddBlock(block)
	b.bcLock.Unlock()

	return nil

}

func (b *blockMgr) validateBlock(block *proto.Block) error {
	err := b.blockValidator.ValidateBlock(block)
	if err != nil {
		logger.Error("区块不合法")
		return err
	}
	for _, signedTransaction := range block.Data.Transactions {
		err := b.txValidator.ValidateTx(signedTransaction)
		if err != nil {
			logger.Error("交易不合法")
			return err
		}
	}
	return nil
}

func (b *blockMgr) GetBlock(number int) *proto.Block {
	b.bcLock.RLock()
	defer b.bcLock.RUnlock()

	return b.bc.GetBlock(number)
}

func (b *blockMgr) GetChain() *chain.BlockChain {
	return b.bc
}

func (b *blockMgr) RegisterBlockHandler(handler BlockHandler) {
	b.handlerLock.Lock()
	defer b.handlerLock.Unlock()

	b.blockHandlers = append(b.blockHandlers, handler)
}

func (b *blockMgr) RegisterTxHandler(handler TxHandler) {
	b.handlerLock.Lock()
	defer b.handlerLock.Unlock()

	txHandlers := b.txHandlers[handler.TxType()]
	b.txHandlers[handler.TxType()] = append(txHandlers, handler)
}
