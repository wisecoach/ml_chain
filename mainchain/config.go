package mainchain

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/proto"
	"time"
)

type Config struct {
	// --------------- crypto config ----------------------------------------------------
	// private key for trainer
	Sk bccsp.Key
	// option for key import
	KeyImportOpts bccsp.KeyImportOpts
	// option for hash
	HashOpts bccsp.HashOpts
	// option for sign
	SignerOpts bccsp.SignerOpts

	// --------------- discovery config -------------------------------------------------
	// peer for trainer
	Self *proto.RemotePeer
	// the bootstrap peers for the task
	BootstrapPeers []*proto.RemotePeer

	// --------------- rpc config -------------------------------------------------------
	// timeout for rpc
	TimeoutRPC time.Duration

	// --------------- chain config -----------------------------------------------------
	GenesisBlock *proto.Block
	ChainId      string

	// --------------- consensus config -------------------------------------------------
	HashInterval      time.Duration
	MaxInterval       time.Duration
	MaxTxNum          int
	NumToConfirm      int
	DefaultDifficulty uint64
}
