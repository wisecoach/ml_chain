package node

import (
	"github.com/wisecoach/ml_chain/comm/comm"
	"time"
)

type Config struct {
	Self           *comm.RemotePeer
	BootstrapPeers []*comm.RemotePeer
	timeoutRPC     time.Duration
}

func NewConfig(self *comm.RemotePeer, bootstrapPeers []*comm.RemotePeer, timeoutRPC time.Duration) *Config {
	return &Config{
		Self:           self,
		BootstrapPeers: bootstrapPeers,
		timeoutRPC:     timeoutRPC,
	}
}
