package peer

import (
	comm2 "github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/discovery"
	"net/rpc"
)

type Peer struct {
	ChainId   string
	comm      comm2.Comm
	discovery discovery.Discovery
}

func New(config comm2.Config, server *rpc.Server) *Peer {
	peer := &Peer{
		comm:      comm2.New(server),
		discovery: nil,
	}

	return peer
}
