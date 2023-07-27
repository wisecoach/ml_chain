package manager

import (
	"errors"
	"fmt"
	"github.com/wisecoach/ml_chain/block/chain"
	"github.com/wisecoach/ml_chain/proto"
	"go.uber.org/zap"
	"reflect"
	"sync"
)

type blockMgr struct {
	bc             *chain.BlockChain
	blockValidator BlockValidator
	txValidator    TxValidator
	blockHandlers  []BlockHandler
	txHandlers     map[reflect.Type][]TxHandler

	logger      *zap.Logger
	handlerLock sync.RWMutex
	bcLock      sync.RWMutex
}

func (b *blockMgr) GetLatestBlock() (*proto.Block, error) {
	if b.bc.Number == 0 {
		return nil, errors.New("now, the blockchain don't have any block")
	}
	return b.GetBlock(b.bc.Number)
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
		b.logger.Error("block is not valid")
		return err
	}
	for i, signedTransaction := range block.Data.Transactions {
		err := b.txValidator.ValidateTx(signedTransaction)
		if err != nil {
			b.logger.Error(fmt.Sprintf("transaction is not valid, index is %d", i))
			return err
		}
	}
	return nil
}

func (b *blockMgr) GetBlock(number int) (*proto.Block, error) {
	b.bcLock.RLock()
	defer b.bcLock.RUnlock()

	block := b.bc.GetBlock(number)
	if block == nil {
		return nil, errors.New(fmt.Sprintf("the height of blockchain is just %d, but you want get %d", b.bc.Number, number))
	}
	return block, nil
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
