package proto

import "time"

type Header struct {
	Creator   []byte // public key
	ChainId   string
	TxId      string
	Timestamp time.Time
}

// Envelope envelope represent a message with a signature and the payload will have the PublicKey for verify the signature
type Envelope[T any] struct {
	Payload   T
	Signature []byte
}
