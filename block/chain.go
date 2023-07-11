package block

type Chain struct {
	Blocks []*Block
}

func New() *Chain {
	chain := &Chain{make([]*Block, 0)}
	return chain
}

func (chain *Chain) AddBlock(newBlock *Block) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock.PrevHash = prevBlock.DataHash
	chain.Blocks = append(chain.Blocks, newBlock)
}

func (chain *Chain) GetBlock(number int) *Block {
	return chain.Blocks[number]
}
