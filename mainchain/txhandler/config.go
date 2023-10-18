package txhandler

import "github.com/wisecoach/ml_chain/bccsp"

type Config struct {
	TrainerNum   int
	Sk           bccsp.Key
	ShardCreator []byte
}
