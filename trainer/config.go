package trainer

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/proto"
	"time"
)

type Config struct {
	// --------------- task config ------------------------------------------------------
	// task id for trainer
	TaskId string
	// index of client
	Cindex string
	// Aggregator
	Aggregator []byte
	// trainer number
	TrainerNum int
	// task genesis block
	GenesisBlock *proto.Block
	// python server api base
	ApiBaseUrl string

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

	// --------------- notary config ----------------------------------------------------
	Notaries []*proto.RemotePeer

	// --------------- rpc config -------------------------------------------------------
	// timeout for rpc
	TimeoutRPC time.Duration

	// --------------- chain config -----------------------------------------------------
	// block number can be resident in memory
	MaxBlockNumInMemory int
}
