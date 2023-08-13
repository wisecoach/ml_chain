package proto

// Block is the block of task shard chain
type Block struct {
	Header *BlockHeader
	Data   *BlockData
}

type BlockHeader struct {
	DataHash    []byte
	PrevHash    []byte
	BlockNumber int
	Miner       []byte
	Nonce       uint32
}

type BlockData struct {
	Transactions []*Envelope[*Transaction]
}
