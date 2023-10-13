package chain

import (
	"errors"
	"fmt"
	"github.com/wisecoach/ml_chain/proto"
)

type BlockChain struct {
	Blocks        []*proto.Block
	Number        int
	indexInMemory int

	config *Config
}

func NewBlockChain(config *Config) *BlockChain {
	chain := &BlockChain{
		Blocks:        make([]*proto.Block, 0),
		Number:        0,
		indexInMemory: 0,
		config:        config,
	}
	return chain
}

func (chain *BlockChain) AddBlock(newBlock *proto.Block) error {
	if len(chain.Blocks) > 0 {
		prevBlock := chain.Blocks[len(chain.Blocks)-1]
		newBlock.Header.PrevHash = prevBlock.Header.DataHash
	} else {
		newBlock.Header.PrevHash = newBlock.Header.DataHash
	}
	if chain.Number >= chain.config.MaxBlockNumInMemory {
		chain.indexInMemory = chain.Blocks[0].Header.BlockNumber + 1
		chain.Blocks = chain.Blocks[1:]
	}
	err := chain.persistBlock(newBlock)
	if err != nil {
		println(err.Error())
		return err
	}
	chain.Blocks = append(chain.Blocks, newBlock)
	chain.Number += 1
	return nil
}

func (chain *BlockChain) GetBlock(number int) (*proto.Block, error) {
	if number < 0 {
		return nil, errors.New("block number cannot be negative")
	}
	if number >= chain.Number {
		return nil, errors.New(fmt.Sprintf("blockchain's height is %d, don't have block %d", chain.Number, number))
	}
	if number < chain.indexInMemory {
		return chain.loadBlock(number)
	}
	actualIndex := number - chain.indexInMemory
	return chain.Blocks[actualIndex], nil
}

func (chain *BlockChain) GetNumber() int {
	return chain.Number
}

func (chain *BlockChain) persistBlock(block *proto.Block) error {
	// marshal, err := json.Marshal(block)
	// if err != nil {
	// 	return err
	// }
	// path := fmt.Sprintf("data/blocks/%s/", chain.config.ChainId)
	// fileName := path + fmt.Sprintf("%d.block", block.Header.BlockNumber)
	// err = os.MkdirAll(path, 0666)
	// if err != nil {
	// 	return err
	// }
	// err = os.WriteFile(fileName, marshal, 0666)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (chain *BlockChain) loadBlock(number int) (*proto.Block, error) {
	// TODO load block from file system
	return nil, errors.New("not implements")
}
