package task

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"time"
)

type taskClient struct {
	config *Config

	mcs    crypto.MessageCryptoService
	node   *node.Node
	logger *zap.Logger
}

func NewTaskClient(config *Config, mcs crypto.MessageCryptoService, node *node.Node) Client {
	return &taskClient{
		config: config,
		mcs:    mcs,
		node:   node,
		logger: log.GetLogger(),
	}
}

func (t *taskClient) CreateTask(task *Task) {
	genesis := task.TaskGenesis
	transaction := &proto.Transaction{
		Parent: nil,
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      uuid.New().String(),
			Timestamp: time.Time{},
		},
		Payload: genesis,
	}
	txBytes, err := json.Marshal(transaction)
	if err != nil {
		t.logger.Error("transaction marshal failed")
		return
	}
	signature, err := t.mcs.Sign(txBytes)
	if err != nil {
		t.logger.Error("transaction sign failed")
		return
	}
	signedTransaction := &proto.Envelope[*proto.Transaction]{
		Payload:   transaction,
		Signature: signature,
	}
	msg := &proto.Message{
		Content: &proto.TransactionMessage{
			Transaction: signedTransaction,
		},
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      "",
			Timestamp: time.Time{},
		},
	}
	t.node.SendToPeers(msg, t.node.Self())
}

func (t *taskClient) FinishTask(task *FinishedTask) {
	transaction := &proto.Transaction{
		Parent: nil,
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      uuid.New().String(),
			Timestamp: time.Time{},
		},
		Payload: task,
	}
	txBytes, err := json.Marshal(transaction)
	if err != nil {
		t.logger.Error("transaction marshal failed")
		return
	}
	signature, err := t.mcs.Sign(txBytes)
	if err != nil {
		t.logger.Error("transaction sign failed")
		return
	}
	signedTransaction := &proto.Envelope[*proto.Transaction]{
		Payload:   transaction,
		Signature: signature,
	}
	msg := &proto.Message{
		Content: &proto.TransactionMessage{
			Transaction: signedTransaction,
		},
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      "",
			Timestamp: time.Time{},
		},
	}
	t.node.SendToPeers(msg, t.node.Self())
}

func (t *taskClient) RegisterManager(pk []byte) {
	transaction := &proto.Transaction{
		Parent: nil,
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      uuid.New().String(),
			Timestamp: time.Time{},
		},
		Payload: &proto.ManagerRegister{Registrar: pk},
	}
	txBytes, err := json.Marshal(transaction)
	if err != nil {
		t.logger.Error("transaction marshal failed")
		return
	}
	signature, err := t.mcs.Sign(txBytes)
	if err != nil {
		t.logger.Error("transaction sign failed")
		return
	}
	signedTransaction := &proto.Envelope[*proto.Transaction]{
		Payload:   transaction,
		Signature: signature,
	}
	msg := &proto.Message{
		Content: &proto.TransactionMessage{
			Transaction: signedTransaction,
		},
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      "",
			Timestamp: time.Time{},
		},
	}
	t.node.SendToPeers(msg, t.node.Self())
}

func (t *taskClient) RevokeManager(pk []byte) {
	transaction := &proto.Transaction{
		Parent: nil,
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      uuid.New().String(),
			Timestamp: time.Time{},
		},
		Payload: &proto.ManagerRevoke{Manager: pk},
	}
	txBytes, err := json.Marshal(transaction)
	if err != nil {
		t.logger.Error("transaction marshal failed")
		return
	}
	signature, err := t.mcs.Sign(txBytes)
	if err != nil {
		t.logger.Error("transaction sign failed")
		return
	}
	signedTransaction := &proto.Envelope[*proto.Transaction]{
		Payload:   transaction,
		Signature: signature,
	}
	msg := &proto.Message{
		Content: &proto.TransactionMessage{
			Transaction: signedTransaction,
		},
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.ChainId,
			TxId:      "",
			Timestamp: time.Time{},
		},
	}
	t.node.SendToPeers(msg, t.node.Self())
}
