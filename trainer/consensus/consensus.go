package consensus

import (
	"encoding/json"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/iteration"
	"github.com/wisecoach/ml_chain/trainer/role"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type TaskConsensus struct {
	logger   zap.Logger
	txChan   chan *proto.Envelope[*proto.Transaction]
	stopChan chan struct{}
	config   *Config

	blockManager     manager.BlockManager
	iterationManager iteration.Manager
	roleManager      role.Manager
	mcs              crypto.MessageCryptoService
	node             *node.Node
}

func (t *TaskConsensus) Order(transaction *proto.Envelope[*proto.Transaction]) {
	current := t.iterationManager.GetIteration()
	amAggregator := t.roleManager.AmAggregator(current)
	// if not aggregator, forward msg to aggregator
	if !amAggregator {
		msg := &proto.Message{
			Content: &proto.TransactionMessage{Transaction: transaction},
			Header: &proto.Header{
				Creator:   nil,
				ChainId:   t.config.TaskId,
				TxId:      "",
				Timestamp: time.Time{},
			},
		}
		aggregator := t.node.Lookup(t.roleManager.GetAggregator(current))
		t.node.SendToPeers(msg, aggregator)
		return
	}
	// if node is aggregator, and transaction is ModelIteration
	modelIteration, isModelIteration := transaction.Payload.Payload.(*proto.ModelIteration)
	if isModelIteration {
		// if current + 1 > transaction.Iteration, means transaction is expired
		if current+1 > modelIteration.Iteration {
			return
		} else if current+1 < modelIteration.Iteration {
			// if current + 1 < transaction.Iteration, it won't occur because of the synchronization of federated learning
			return
		} else {
			t.txChan <- transaction
		}
	}
}

func (t *TaskConsensus) Consensus(block *proto.Envelope[*proto.Block]) {
	// check signature
	miner := block.Payload.Header.Miner
	blockBytes, err := json.Marshal(block.Signature)
	if err != nil {
		t.logger.Error("block json marshal failed" + err.Error())
		return
	}
	_, err = t.mcs.Verify(miner, blockBytes, block.Signature)
	if err != nil {
		t.logger.Error("block signature is invalid" + err.Error())
		return
	}
	latestBlock, err := t.blockManager.GetLatestBlock()
	if err != nil {
		return
	}
	// check prev hash and data hash
	// TODO implement a machine to cache orphan block
	if !reflect.DeepEqual(latestBlock.Header.DataHash, block.Payload.Header.PrevHash) {
		t.logger.Error("block is a orphan block")
		return
	}
	// check if miner is aggregator
	current := t.iterationManager.GetIteration()
	aggregator := t.node.Lookup(t.roleManager.GetAggregator(current))
	if !reflect.DeepEqual(aggregator, miner) {
		t.logger.Error("miner of block is not aggregator")
		return
	}
	// confirm block to blockManager
	err = t.blockManager.ConfirmBlock(block.Payload)
	if err != nil {
		t.logger.Error("confirm block failed" + err.Error())
	}
}

func (t *TaskConsensus) Start() {
	for {
		select {
		case <-t.stopChan:
			return
		case tx := <-t.txChan:
			t.processTransaction(tx)
		}
	}
}

func (t *TaskConsensus) processTransaction(transaction *proto.Envelope[*proto.Transaction]) {
	block, err := t.createBlock(transaction)
	if err != nil {
		return
	}
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return
	}
	signature, err := t.mcs.Sign(blockBytes)
	if err != nil {
		return
	}
	envelop := &proto.Envelope[*proto.Block]{
		Payload:   block,
		Signature: signature,
	}
	blockMsg := &proto.Message{
		Content: &proto.BlockMessage{Block: envelop},
		Header: &proto.Header{
			Creator:   t.node.Self().PublicKey,
			ChainId:   t.config.TaskId,
			TxId:      "",
			Timestamp: time.Time{},
		},
	}
	t.node.SendWithFilter(blockMsg, func(peer *comm.RemotePeer) bool { return true })
}

func (t *TaskConsensus) createBlock(txs ...*proto.Envelope[*proto.Transaction]) (*proto.Block, error) {
	blockData := &proto.BlockData{
		Transactions: txs,
	}
	dataBytes, err := json.Marshal(blockData)
	if err != nil {
		t.logger.Error("blockData json marshal")
		return nil, err
	}
	dataHash, err := t.mcs.Hash(dataBytes)
	latestBlock, err := t.blockManager.GetLatestBlock()
	if err != nil {
		return nil, err
	}
	block := &proto.Block{
		Header: &proto.BlockHeader{
			DataHash:    dataHash,
			PrevHash:    latestBlock.Header.DataHash,
			BlockNumber: 0,
			Miner:       t.node.Self().PublicKey,
		},
		Data: blockData,
	}
	return block, nil
}
