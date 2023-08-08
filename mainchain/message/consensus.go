package message

import (
	"github.com/wisecoach/ml_chain/block/consensus"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
)

type BlockMessageListener struct {
	consensus consensus.Consensus
}

func NewBlockMessageListener(consensus consensus.Consensus) node.MessageListener {
	return &BlockMessageListener{consensus: consensus}
}

func (b *BlockMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	block := message.Envelope.Payload.GetBlock()
	b.consensus.Consensus(block.Block)
}

type TransactionMessageListener struct {
	consensus consensus.Consensus
}

func NewTransactionMessageListener(consensus consensus.Consensus) node.MessageListener {
	return &TransactionMessageListener{consensus: consensus}
}

func (t *TransactionMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	transaction := message.Envelope.Payload.GetTransaction()
	t.consensus.Order(transaction.Transaction)
}
