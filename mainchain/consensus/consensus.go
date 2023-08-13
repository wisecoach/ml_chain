package consensus

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/wisecoach/ml_chain/block/consensus"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"time"
)

type Branch struct {
	tipNum int
	tip    *proto.Block
	blocks []*proto.Block
	txs    map[string]*proto.Envelope[*proto.Transaction]
}

func New(config *Config, blockManager manager.BlockManager, node *node.Node, mcs crypto.MessageCryptoService) consensus.Consensus {
	return &MainChainConsensus{
		blockManager:     blockManager,
		node:             node,
		mcs:              mcs,
		longestBranch:    nil,
		branches:         make([]*Branch, 0),
		branchLock:       sync.RWMutex{},
		orphanBlocks:     make(map[string]*proto.Block),
		orphanLock:       sync.RWMutex{},
		txPool:           make(map[string]*proto.Envelope[*proto.Transaction]),
		enoughTxChan:     make(chan struct{}),
		poolLock:         sync.RWMutex{},
		stopPoWChan:      make(chan struct{}),
		changeBranchChan: make(chan struct{}),
		lock:             sync.RWMutex{},
		config:           config,
		logger:           log.GetLogger(),
	}
}

type MainChainConsensus struct {
	blockManager manager.BlockManager
	node         *node.Node
	mcs          crypto.MessageCryptoService

	longestBranch *Branch
	branches      []*Branch
	branchLock    sync.RWMutex
	orphanBlocks  map[string]*proto.Block // prevHash to orphanBlock
	orphanLock    sync.RWMutex
	txPool        map[string]*proto.Envelope[*proto.Transaction]
	enoughTxChan  chan struct{}
	poolLock      sync.RWMutex

	stopPoWChan      chan struct{}
	changeBranchChan chan struct{}

	lock   sync.RWMutex // lock for all resource
	config *Config
	logger *zap.Logger
}

func (m *MainChainConsensus) Order(transaction *proto.Envelope[*proto.Transaction]) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.txPool[transaction.Payload.Header.TxId] = transaction
	if len(m.txPool) >= m.config.MaxTxNum {
		m.enoughTxChan <- struct{}{}
	}
}

func (m *MainChainConsensus) Consensus(block *proto.Envelope[*proto.Block]) {
	err := m.validateBlock(block)
	if err != nil {
		m.logger.Error(fmt.Sprintf("validate block %d failed", block.Payload.Header.BlockNumber))
		return
	}
	m.consensus(block.Payload)
}

// consensus
//
//	@Description: consensus the validated block
func (m *MainChainConsensus) consensus(block *proto.Block) {
	blockNum := block.Header.BlockNumber
	confirmHeight := m.blockManager.GetHeight()
	// if block is older than confirm block
	if blockNum <= confirmHeight {
		m.logger.Debug(fmt.Sprintf("receive old block %d, but height of blockchain is %d", blockNum, confirmHeight))
		return
	}

	// if block is newer than longest branch + 1, it must be a orphan block
	if blockNum > m.longestBranch.tipNum+1 {
		m.addOrphan(block)
	}

	// if block is next block of the branch, append to branch
	m.lock.RLock()
	index := -1
	for i, branch := range m.branches {
		if reflect.DeepEqual(branch.tip.Header.DataHash, block.Header.PrevHash) {
			index = i
		}
	}
	var branch *Branch
	if index != -1 {
		branch = m.branches[index]
	}
	m.lock.RUnlock()

	if branch != nil {
		m.appendToBranch(branch, block)
		return
	}
	// now, block may be orphan block, a new branch or a duplicated block

	// check block is a duplicate block or a new branch
	distance := blockNum - confirmHeight
	m.lock.RLock()
	branches := m.branches
	m.lock.RUnlock()

	for _, branch := range branches {
		if branch.tipNum >= blockNum {
			if reflect.DeepEqual(branch.blocks[distance].Header.DataHash, block.Header.DataHash) {
				// block is duplicated block
				m.logger.Debug(fmt.Sprintf("receive duplicated block %d", blockNum))
				return
			} else if reflect.DeepEqual(block.Header.PrevHash, branch.blocks[distance-1]) {
				// block is new branch
				m.logger.Debug(fmt.Sprintf("receive a block %d belong to a new branch", blockNum))
				// TODO need deep copy?
				m.newBranch(branch.blocks[:distance-1], block)
			} else {
				// block is orphan block
				m.logger.Debug(fmt.Sprintf("receive a orphan block %d", blockNum))
				m.addOrphan(block)
			}
		} else {
			// branch.tipNum < blockNum, block is orphan
			m.addOrphan(block)
			return
		}
	}

}

// appendToBranch
//
//	@Description: append block to branch
//		1. if the branch length longer than other, it will be longest branch, and when length is equal to other, whose datahash is smaller will be longest one
//		2. if branch is longer than NumToConfirm, confirm the branch first block, and abandon branch whose first block with equal number but different datahash
//	 	3. if block is the prev block of orphan block, consensus the orphan block
func (m *MainChainConsensus) appendToBranch(branch *Branch, block *proto.Block) {
	m.branchLock.Lock()
	// append block
	branch.blocks = append(branch.blocks, block)
	branch.tip = block
	branch.tipNum += 1
	for _, tx := range block.Data.Transactions {
		branch.txs[tx.Payload.Header.TxId] = tx
	}

	// check if need to change longest branch
	if branch != m.longestBranch && branch.tipNum >= m.longestBranch.tipNum {
		if branch.tipNum == m.longestBranch.tipNum {
			if m.longestBranch.tipNum == block.Header.BlockNumber && string(m.longestBranch.tip.Header.DataHash) > string(block.Header.DataHash) {
				m.changeLongestBranch(branch)
			}
		} else {
			m.changeLongestBranch(branch)
		}
	}
	m.branchLock.Unlock()

	// check if length is enough to confirm
	branchLen := len(branch.blocks)
	if branchLen > m.config.NumToConfirm {
		m.confirm(branch.blocks[0])
	}

	// check if block is the prev block of orphan block
	m.orphanLock.RLock()
	dataHash := base64.StdEncoding.EncodeToString(block.Header.DataHash)
	orphan, exists := m.orphanBlocks[dataHash]
	m.orphanLock.RUnlock()
	if exists {
		m.consensus(orphan)
	}
}

// addOrphan
//
//	@Description: save orphan block
func (m *MainChainConsensus) addOrphan(block *proto.Block) {
	hash := base64.StdEncoding.EncodeToString(block.Header.PrevHash)
	m.orphanLock.Lock()
	m.orphanBlocks[hash] = block
	m.orphanLock.Unlock()
}

// newBranch
//
//	@Description: create a new branch, it may be longest branch, when length is equal to old branch, and newBlock.Header.DataHash is smaller than the old
func (m *MainChainConsensus) newBranch(prevBlocks []*proto.Block, newBlock *proto.Block) {
	blocks := prevBlocks
	blocks = append(blocks, newBlock)
	newBranch := &Branch{
		tipNum: newBlock.Header.BlockNumber,
		tip:    newBlock,
		blocks: blocks,
		txs:    make(map[string]*proto.Envelope[*proto.Transaction]),
	}
	for _, block := range blocks {
		for _, tx := range block.Data.Transactions {
			newBranch.txs[tx.Payload.Header.TxId] = tx
		}
	}

	m.branchLock.Lock()
	m.branches = append(m.branches, newBranch)
	// check if new branch is longest branch
	if m.longestBranch.tipNum == newBlock.Header.BlockNumber && string(m.longestBranch.tip.Header.DataHash) > string(newBlock.Header.DataHash) {
		m.changeLongestBranch(newBranch)
	}
	m.branchLock.Unlock()
}

// changeLongestBranch
//
//	@Description: change the longest branch, it will restart pow
//				Note: it's will be called synchronized, no need to lock
func (m *MainChainConsensus) changeLongestBranch(branch *Branch) {
	m.longestBranch = branch
	m.changeBranchChan <- struct{}{}
}

func (m *MainChainConsensus) confirm(block *proto.Block) {
	m.lock.Lock()
	newBranches := make([]*Branch, 0)

	// remove branch without same first block, and pop first block and it's txs
	for _, branch := range m.branches {
		if reflect.DeepEqual(branch.blocks[0].Header.DataHash, block.Header.DataHash) {
			for _, tx := range block.Data.Transactions {
				delete(branch.txs, tx.Payload.Header.TxId)
			}
			newBranches = append(newBranches, branch)
		}
	}
	// remove duplicated tx in pool
	for _, tx := range block.Data.Transactions {
		delete(m.txPool, tx.Payload.Header.TxId)
	}

	m.lock.Unlock()
}

func (m *MainChainConsensus) validateBlock(block *proto.Envelope[*proto.Block]) error {
	// TODO implement me
	return nil
}

func (m *MainChainConsensus) Start() {
	go m.start()
}

func (m *MainChainConsensus) start() {
	for {
		var block *proto.Block

		select {
		case <-time.After(m.config.MaxInterval):
			block = m.createBlock()
		case <-m.enoughTxChan:
			block = m.createBlock()
		case <-m.changeBranchChan:
			m.stopPoWChan <- struct{}{}
		}

		// if create block successfully, send to other peer
		if block != nil {
			blockBytes, err := json.Marshal(block)
			if err != nil {
				return
			}
			signature, err := m.mcs.Sign(blockBytes)
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
					Creator:   m.node.Self().PublicKey,
					ChainId:   m.config.ChainId,
					TxId:      "",
					Timestamp: time.Time{},
				},
			}
			err = m.blockManager.ConfirmBlock(block)
			if err != nil {
				m.logger.Error("confirm block failed, err: " + err.Error())
				return
			}
			m.node.SendWithFilter(blockMsg, func(peer *proto.RemotePeer) bool { return true })
		}
	}
}

func (m *MainChainConsensus) createBlock() *proto.Block {
	m.lock.Lock()
	blockNum := m.longestBranch.tipNum + 1
	prevHash := m.longestBranch.tip.Header.DataHash
	transactions := make([]*proto.Envelope[*proto.Transaction], 0)
	for txId, tx := range m.txPool {
		// don't exceed maxTxNum and transaction in branch, when selected, remove from txPool
		if _, exists := m.longestBranch.txs[txId]; !exists && len(transactions) < m.config.MaxTxNum {
			transactions = append(transactions, tx)
			delete(m.txPool, txId)
		}
	}
	dataBytes, err := json.Marshal(transactions)
	if err != nil {
		return nil
	}
	dataHash, err := m.mcs.Hash(dataBytes)
	if err != nil {
		return nil
	}

	block := &proto.Block{
		Header: &proto.BlockHeader{
			DataHash:    dataHash,
			PrevHash:    prevHash,
			BlockNumber: blockNum,
			Miner:       m.node.Self().PublicKey,
			Nonce:       0,
		},
		Data: &proto.BlockData{
			Transactions: transactions,
		},
	}
	m.lock.Unlock()

	blockChan := m.pow(block)
	minedBlock := <-blockChan
	m.poolLock.Lock()
	// when mining failed, revert the transactions
	if minedBlock == nil {
		for _, tx := range block.Data.Transactions {
			m.txPool[tx.Payload.Header.TxId] = tx
		}
	}
	m.poolLock.Unlock()

	return minedBlock
}

// pow
//
//	@Description: PoW, try to find and pad the nonce of block
//	@param block block to mine
//	@return <-chan block with nonce, nil means it's stopped
func (m *MainChainConsensus) pow(block *proto.Block) <-chan *proto.Block {
	blockChan := make(chan *proto.Block)
	cnt := uint32(0)
	go func() {
		for {
			// init difficulty
			difficulty := m.config.DefaultDifficulty * uint64(len(block.Data.Transactions))
			select {
			case <-m.stopPoWChan:
				blockChan <- nil
			case <-time.After(m.config.HashInterval):
				block.Header.Nonce = cnt
				bytes, err := json.Marshal(block)
				if err != nil {
					return
				}
				hash, err := m.mcs.Hash(bytes)
				if err != nil {
					return
				}
				if binary.LittleEndian.Uint64(hash) < difficulty {
					blockChan <- block
					return
				}
			}
		}
	}()
	return blockChan
}
