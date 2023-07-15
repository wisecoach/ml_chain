package data

import "time"

type SignedTransaction struct {
	Payload   *Transaction
	Signature []byte
}

// Transaction represent the operation for state machine
type Transaction struct {
	Parent    *Block
	Header    *Header
	Payload   *Payload
	Id        string
	ChainId   string
	Timestamp *time.Time
}
