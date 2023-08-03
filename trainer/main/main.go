package main

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer"
	"github.com/wisecoach/ml_chain/util/log"
	"strconv"
	"time"
)

func main() {

	csp, _ := sw.NewBCCSP()
	trainerNumber := 2
	bootstrapPeers := make([]*proto.RemotePeer, 0)
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
		bootstrapPeers = append(bootstrapPeers, &proto.RemotePeer{
			Endpoint:  "127.0.0.1:" + strconv.Itoa(i+10000),
			PublicKey: pkBytes,
		})
	}

	transactions := make([]*proto.Envelope[*proto.Transaction], 0)
	taskTransaction := &proto.Transaction{
		Parent: nil,
		Header: nil,
		Payload: &proto.TaskGenesis{
			ModelStructure: nil,
			ManagerList:    [][]byte{bootstrapPeers[0].PublicKey},
			InitWeight: &proto.Envelope[*proto.GlobalWeight]{
				Payload: &proto.GlobalWeight{
					Iteration:    0,
					WeightVector: nil,
					Aggregator:   nil,
				},
				Signature: nil,
			},
		},
	}
	transactions = append(transactions, &proto.Envelope[*proto.Transaction]{
		Payload:   taskTransaction,
		Signature: nil,
	})

	genesis := &proto.Block{
		Header: &proto.BlockHeader{
			DataHash:    nil,
			PrevHash:    nil,
			BlockNumber: 0,
			Miner:       nil,
		},
		Data: &proto.BlockData{
			Transactions: transactions,
		},
	}

	for i := 0; i < trainerNumber; i++ {
		config := &trainer.Config{
			TaskId:         "task1",
			ValidatorNum:   1,
			GenesisBlock:   genesis,
			ApiBaseUrl:     "http://localhost:8999",
			Sk:             privateKeys[i],
			KeyImportOpts:  &bccsp.ECDSAPrivateKeyImportOpts{Temporary: false},
			HashOpts:       &bccsp.SHA256Opts{},
			SignerOpts:     nil,
			Self:           bootstrapPeers[i],
			BootstrapPeers: bootstrapPeers,
			TimeoutRPC:     time.Second * 1000,
		}
		go trainer.New(config)
	}

	select {
	case <-time.After(time.Second * 3600):
	}

}
