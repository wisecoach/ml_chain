package data

// Block is the block of task shard chain
type Block struct {
	DataHash     []byte
	PrevHash     []byte
	BlockNumber  int
	Transactions []*SignedTransaction
}
