package trainer

import (
	"github.com/wisecoach/ml_chain/block/consensus"
	"github.com/wisecoach/ml_chain/block/data"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/trainer/role"
	"net/rpc"
)

type Trainer struct {
	roleManager  role.Manager
	blockManager manager.BlockManager
	consensus    consensus.Consensus
	blockchain   *data.BlockChain
	node         *node.Node
	server       *rpc.Server
}

func New(config *Config) *Trainer {
	t := &Trainer{}
	t.server = rpc.NewServer()
	t.blockchain = data.NewBlockChain()

	t.node = node.New(
		node.NewConfig(config.Self, config.BootstrapPeers, config.timeoutRPC),
		t.server)

	return t
}
