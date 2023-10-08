package manager

import (
	"errors"
	"fmt"
	"github.com/wisecoach/ml_chain/block/chain"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/proto"
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

func New(bc *chain.BlockChain, mcs crypto.MessageCryptoService) BlockManager {
	blockManager := &blockMgr{
		bc:             bc,
		blockValidator: NewBlockValidator(mcs),
		txValidator:    NewTxValidator(mcs),
		blockHandlers:  make([]BlockHandler, 0),
		txHandlers:     make(map[reflect.Type][]TxHandler),
		handlerLock:    sync.RWMutex{},
		bcLock:         sync.RWMutex{},
	}
	return blockManager
}

func (b *blockMgr) GetHeight() int {
	b.bcLock.RLock()
	defer b.bcLock.RUnlock()

	return b.bc.Number
}

func (b *blockMgr) GetLatestBlock() (*proto.Block, error) {
	if b.bc.Number == 0 {
		return nil, errors.New("now, the blockchain don't have any block")
	}
	return b.GetBlock(b.bc.Number - 1)
}

func (b *blockMgr) ConfirmBlock(block *proto.Block) error {
	// validate the block
	err := b.blockValidator.ValidateBlock(block)
	if err != nil {
		return err
	}

	b.bcLock.Lock()
	// add block to blockchain
	b.bc.AddBlock(block)

	b.bcLock.Unlock()

	// handle the block and txs in the block
	// TODO if need to deep copy the array to avoid long time locking
	b.handlerLock.RLock()
	blockHandlers := b.blockHandlers
	txHandlers := b.txHandlers

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

	b.handlerLock.RUnlock()

	return nil

}

func (b *blockMgr) GetBlock(number int) (*proto.Block, error) {
	b.bcLock.RLock()
	defer b.bcLock.RUnlock()

	block, err := b.bc.GetBlock(number)
	if err != nil {
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

func (b *blockMgr) validateBlock(block *proto.Block) error {
	err := b.blockValidator.ValidateBlock(block)
	if err != nil {
		return err
	}
	for _, signedTransaction := range block.Data.Transactions {
		err := b.txValidator.ValidateTx(signedTransaction)
		if err != nil {
			return err
		}
	}
	return nil
}
