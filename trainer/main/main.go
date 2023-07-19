package main

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/trainer"
	"github.com/wisecoach/ml_chain/util/logger"
	"strconv"
	"time"
)

func main() {
	csp, _ := sw.NewBCCSP()
	trainerNumber := 2
	bootstrapPeers := make([]*comm.RemotePeer, 0)
	logger.Debug("sss")
	for i := 0; i < trainerNumber; i++ {
		key, _ := csp.KeyGen(&bccsp.AES256KeyGenOpts{
			Temporary: false,
		})
		pk := key.SKI()
		bootstrapPeers = append(bootstrapPeers, &comm.RemotePeer{
			Endpoint:  "127.0.0.1:" + strconv.Itoa(i+10000),
			PublicKey: pk,
		})
	}

	for i := 0; i < trainerNumber; i++ {
		config := &trainer.Config{
			ChainId:        "task1",
			Self:           bootstrapPeers[i],
			BootstrapPeers: bootstrapPeers,
			KeyImportOpts:  &bccsp.AES256ImportKeyOpts{Temporary: false},
			HashOpts:       &bccsp.SHA256Opts{},
			SignerOpts:     nil,
		}
		trainer.New(config)
	}

	select {
	case <-time.After(time.Second * 5):
	}

}
