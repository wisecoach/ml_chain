package main

import (
	"encoding/json"
	"fmt"
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/mainchain"
	"github.com/wisecoach/ml_chain/mainchain/task"
	"github.com/wisecoach/ml_chain/proto"
	"os"
	"strconv"
	"time"
)

func main() {
	trainerNum := 40
	validatorNum := 20
	configFile, err := os.Open("data/init_global_weight.json")
	if err != nil {
		return
	}
	jsonParser := json.NewDecoder(configFile)
	var initWeight []float64
	if err = jsonParser.Decode(&initWeight); err != nil {
	}

	csp, _ := sw.NewBCCSP()
	peerNumber := 5
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
			Endpoint:  "127.0.0.1:" + strconv.Itoa(i+9000),
			PublicKey: pkBytes,
		})
	}

	genesis := &proto.Block{
		Header: &proto.BlockHeader{
			DataHash:    []byte{1, 2, 3, 4},
			PrevHash:    nil,
			BlockNumber: 0,
			Miner:       nil,
		},
		Data: &proto.BlockData{
			Transactions: nil,
		},
	}

	for i := 0; i < peerNumber; i++ {
		i := i
		go func() {
			config := &mainchain.Config{
				TrainerNum:          trainerNum,
				ValidatorNum:        validatorNum,
				NumSharePy:          5,
				Sk:                  privateKeys[i],
				KeyImportOpts:       &bccsp.ECDSAPKIXPublicKeyImportOpts{Temporary: false},
				HashOpts:            &bccsp.SHA256Opts{},
				SignerOpts:          nil,
				Self:                bootstrapPeers[i],
				BootstrapPeers:      bootstrapPeers,
				Notaries:            make([]*proto.RemotePeer, 0),
				TimeoutRPC:          time.Second * 1000,
				GenesisBlock:        genesis,
				ChainId:             "main",
				MaxBlockNumInMemory: 10,
				HashInterval:        time.Millisecond * 10,
				MaxInterval:         time.Millisecond * 10,
				MaxTxNum:            1,
				NumToConfirm:        3,
				DefaultDifficulty:   1 << 60,
			}
			peer := mainchain.New(config)
			peer.RegisterManager(bootstrapPeers[i].PublicKey)
			select {
			case <-time.After(time.Second * 2):
			}
			if i == 0 {
				peer.CreateTask(&task.Task{TaskGenesis: &proto.TaskGenesis{
					TaskId: fmt.Sprintf("task%d", i),
					ModelStructure: &proto.ModelStructure{
						Dataset:      "mnist",
						NumClasses:   10,
						Agent:        1,
						TrainerNum:   trainerNum,
						ValidatorNum: validatorNum,
						LearningRate: 0.01,
						Momentum:     0.5,
						Dp:           true,
						DpEpsilon:    0.2,
						DpEpsilon1:   0.2,
						DpDelta:      1e-5,
						DpClip:       300,
						BatchSize:    64,
						Round:        100,
						Lambda:       1,
					},
					InitWeight: &proto.Envelope[*proto.GlobalWeight]{
						Payload: &proto.GlobalWeight{
							WeightVector: initWeight,
						},
						Signature: nil,
					},
				}})
			}
		}()
	}

	select {
	case <-time.After(time.Second * 3600):
	}

}
