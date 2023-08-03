package manager

import "github.com/wisecoach/ml_chain/proto"

func NewBlockValidator() BlockValidator {
	return &blockValidatorImpl{}
}

func NewTxValidator() TxValidator {
	return &txValidatorImpl{}
}

type blockValidatorImpl struct {
}

func (b *blockValidatorImpl) ValidateBlock(block *proto.Block) error {
	return nil
}

type txValidatorImpl struct {
}

func (t *txValidatorImpl) ValidateTx(tx *proto.Envelope[*proto.Transaction]) error {
	return nil
}
