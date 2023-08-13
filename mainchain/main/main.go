package main

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/mainchain"
	"github.com/wisecoach/ml_chain/proto"
	"strconv"
	"time"
)

func main() {
	csp, _ := sw.NewBCCSP()
	peerNumber := 2
	bootstrapPeers := make([]*proto.RemotePeer, 0)
	privateKeys := make([]bccsp.Key, 0)
	for i := 0; i < peerNumber; i++ {
		sk, _ := csp.KeyGen(&bccsp.ECDSAP256KeyGenOpts{
			Temporary: false,
		})
		publicKey, err := sk.PublicKey()
		pkBytes, err := publicKey.Bytes()
		if err != nil {
			return
		}
		privateKeys = append(privateKeys, sk)
		bootstrapPeers = append(bootstrapPeers, &proto.RemotePeer{
			Endpoint:  "127.0.0.1:" + strconv.Itoa(i+11100),
			PublicKey: pkBytes,
		})
	}

	genesis := &proto.Block{
		Header: &proto.BlockHeader{
			DataHash:    nil,
			PrevHash:    nil,
			BlockNumber: 0,
			Miner:       nil,
		},
		Data: &proto.BlockData{
			Transactions: nil,
		},
	}

	for i := 0; i < peerNumber; i++ {
		config := &mainchain.Config{
			Sk:                privateKeys[i],
			KeyImportOpts:     &bccsp.ECDSAPKIXPublicKeyImportOpts{Temporary: false},
			HashOpts:          &bccsp.SHA256Opts{},
			SignerOpts:        nil,
			Self:              bootstrapPeers[i],
			BootstrapPeers:    bootstrapPeers,
			TimeoutRPC:        time.Second * 1000,
			GenesisBlock:      genesis,
			ChainId:           "main",
			HashInterval:      time.Millisecond * 1,
			MaxInterval:       time.Second * 1,
			MaxTxNum:          3,
			NumToConfirm:      3,
			DefaultDifficulty: 1 << 20,
		}
		go mainchain.New(config)
	}

	select {
	case <-time.After(time.Second * 3600):
	}

}
