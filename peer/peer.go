package peer

import (
	"github.com/wisecoach/ml_chain/comm"
	"github.com/wisecoach/ml_chain/comm/discovery"
	"net/rpc"
)

type Peer struct {
	ChainId   string
	comm      comm.Comm
	discovery discovery.Discovery
}

func New(config comm.Config, server *rpc.Server) *Peer {
	peer := &Peer{
		comm:      comm.New(server),
		discovery: nil,
	}

	return peer
}
