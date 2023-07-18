package trainer

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/comm/comm"
	"time"
)

type Config struct {
	ChainId        string
	Self           *comm.RemotePeer
	BootstrapPeers []*comm.RemotePeer
	timeoutRPC     time.Duration
	KeyImportOpts  bccsp.KeyImportOpts
	HashOpts       bccsp.HashOpts
	SignerOpts     bccsp.SignerOpts
}
