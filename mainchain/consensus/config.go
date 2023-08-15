package consensus

import (
	"github.com/wisecoach/ml_chain/proto"
	"time"
)

type Config struct {
	ChainId           string
	HashInterval      time.Duration
	MaxInterval       time.Duration
	MaxTxNum          int
	NumToConfirm      int
	DefaultDifficulty uint64
	GenesisBlock      *proto.Block
}
