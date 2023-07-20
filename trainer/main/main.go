package main

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/trainer"
	"github.com/wisecoach/ml_chain/util/log"
	"strconv"
	"time"
)

func main() {

	csp, _ := sw.NewBCCSP()
	trainerNumber := 2
	bootstrapPeers := make([]*comm.RemotePeer, 0)
	privateKeys := make([]bccsp.Key, 0)
	logger := log.GetLogger()
	logger.Debug("sss")
	for i := 0; i < trainerNumber; i++ {
		sk, _ := csp.KeyGen(&bccsp.ECDSAP256KeyGenOpts{
			Temporary: false,
		})
		publicKey, err := sk.PublicKey()
		pkBytes, err := publicKey.Bytes()
		if err != nil {
			return
		}
		privateKeys = append(privateKeys, sk)
		bootstrapPeers = append(bootstrapPeers, &comm.RemotePeer{
			Endpoint:  "127.0.0.1:" + strconv.Itoa(i+10000),
			PublicKey: pkBytes,
		})
	}

	for i := 0; i < trainerNumber; i++ {
		config := &trainer.Config{
			ChainId:        "task1",
			Sk:             privateKeys[i],
			Self:           bootstrapPeers[i],
			BootstrapPeers: bootstrapPeers,
			TimeoutRPC:     time.Second * 1000,
			KeyImportOpts:  &bccsp.ECDSAPrivateKeyImportOpts{Temporary: false},
			HashOpts:       &bccsp.SHA256Opts{},
			SignerOpts:     nil,
		}
		trainer.New(config)
	}

	select {
	case <-time.After(time.Second * 3600):
	}

}
