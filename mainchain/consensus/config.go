package consensus

import "time"

type Config struct {
	ChainId           string
	HashInterval      time.Duration
	MaxInterval       time.Duration
	MaxTxNum          int
	NumToConfirm      int
	DefaultDifficulty uint64
}
