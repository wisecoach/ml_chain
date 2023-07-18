package node

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/comm/comm"
	"time"
)

type Config struct {
	Self           *comm.RemotePeer
	BootstrapPeers []*comm.RemotePeer
	TimeoutRPC     time.Duration
	KeyImportOpts  bccsp.KeyImportOpts
	HashOpts       bccsp.HashOpts
	SignerOpts     bccsp.SignerOpts
}

func NewConfig(self *comm.RemotePeer, bootstrapPeers []*comm.RemotePeer, timeoutRPC time.Duration,
	keyImportOpts bccsp.KeyImportOpts, hashOpts bccsp.HashOpts, signerOpts bccsp.SignerOpts,
) *Config {
	return &Config{
		Self:           self,
		BootstrapPeers: bootstrapPeers,
		TimeoutRPC:     timeoutRPC,
		KeyImportOpts:  keyImportOpts,
		HashOpts:       hashOpts,
		SignerOpts:     signerOpts,
	}
}
