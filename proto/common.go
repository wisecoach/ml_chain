package proto

import "time"

type Header struct {
	Creator   []byte // public key
	ChainId   string
	TxId      []byte
	Timestamp time.Time
}

type Envelope[T any] struct {
	Payload   T
	signature []byte
}
