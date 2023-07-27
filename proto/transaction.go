package proto

// Transaction represent the operation for state machine
type Transaction struct {
	Parent  *Block
	Header  *Header
	Payload Payload
}
