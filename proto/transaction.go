package proto

// Transaction represent the operation for state machine
type Transaction struct {
	Parent      *Block
	Header      *Header
	Payload     Payload
	NotarySigns [][]byte // not nil only if transaction is a cross transaction
}
