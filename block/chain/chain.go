package chain

import "github.com/wisecoach/ml_chain/proto"

type BlockChain struct {
	Blocks []*proto.Block
	Number int
}

func NewBlockChain() *BlockChain {
	chain := &BlockChain{Blocks: make([]*proto.Block, 0), Number: 0}
	return chain
}

func (chain *BlockChain) AddBlock(newBlock *proto.Block) {
	if len(chain.Blocks) > 0 {
		prevBlock := chain.Blocks[len(chain.Blocks)-1]
		newBlock.Header.PrevHash = prevBlock.Header.DataHash
	} else {
		newBlock.Header.PrevHash = newBlock.Header.DataHash
	}
	chain.Blocks = append(chain.Blocks, newBlock)
	chain.Number += 1
}

func (chain *BlockChain) GetBlock(number int) *proto.Block {
	return chain.Blocks[number]
}

func (chain *BlockChain) GetNumber() int {
	return chain.Number
}
