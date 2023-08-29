package txhandler

import "github.com/wisecoach/ml_chain/bccsp"

type Config struct {
	TrainerNum   int
	ValidatorNum int
	NumSharePy   int
	Sk           bccsp.Key
}
