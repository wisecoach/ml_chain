package trainer

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/comm/comm"
	"time"
)

type Config struct {
	ChainId        string
	Sk             bccsp.Key
	Self           *comm.RemotePeer
	BootstrapPeers []*comm.RemotePeer
	TimeoutRPC     time.Duration
	KeyImportOpts  bccsp.KeyImportOpts
	HashOpts       bccsp.HashOpts
	SignerOpts     bccsp.SignerOpts
}
