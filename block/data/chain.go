package data

type BlockChain struct {
	Blocks []*Block
	Number int
}

func NewBlockChain() *BlockChain {
	chain := &BlockChain{Blocks: make([]*Block, 0), Number: 0}
	return chain
}

func (chain *BlockChain) AddBlock(newBlock *Block) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock.PrevHash = prevBlock.DataHash
	chain.Blocks = append(chain.Blocks, newBlock)
	chain.Number += 1
}

func (chain *BlockChain) GetBlock(number int) *Block {
	return chain.Blocks[number]
}

func (chain *BlockChain) GetNumber() int {
	return chain.Number
}
