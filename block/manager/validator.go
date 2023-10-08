package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/proto"
)

func NewBlockValidator(mcs crypto.MessageCryptoService) BlockValidator {
	return &blockValidatorImpl{mcs: mcs}
}

func NewTxValidator(mcs crypto.MessageCryptoService) TxValidator {
	return &txValidatorImpl{mcs: mcs}
}

type blockValidatorImpl struct {
	mcs crypto.MessageCryptoService
}

func (b *blockValidatorImpl) ValidateBlock(block *proto.Block) error {
	marshal, err := json.Marshal(block.Data)
	if err != nil {
		return err
	}
	hash, err := b.mcs.Hash(marshal)
	if err != nil {
		return err
	}
	if !bytes.Equal(block.Header.DataHash, hash) {
		return errors.New("data hash of the block is not match")
	}
	return nil
}

type txValidatorImpl struct {
	mcs crypto.MessageCryptoService
}

func (t *txValidatorImpl) ValidateTx(tx *proto.Envelope[*proto.Transaction]) error {
	marshal, err := json.Marshal(tx.Payload)
	if err != nil {
		return err
	}
	_, err = t.mcs.Verify(tx.Payload.Header.Creator, marshal, tx.Signature)
	if err != nil {
		return err
	}
	return nil
}
