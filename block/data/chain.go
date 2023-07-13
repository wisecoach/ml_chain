package data

type BlockChain struct {
	Blocks []*Block
}

func New() *BlockChain {
	chain := &BlockChain{make([]*Block, 0)}
	return chain
}

func (chain *BlockChain) AddBlock(newBlock *Block) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock.PrevHash = prevBlock.DataHash
	chain.Blocks = append(chain.Blocks, newBlock)
}

func (chain *BlockChain) GetBlock(number int) *Block {
	return chain.Blocks[number]
}
