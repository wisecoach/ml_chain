package main

import (
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
	testArg := os.Args[1]
	testTime := os.Args[2]
	switch testArg {
	case "1":
		minute, err := strconv.Atoi(testTime)
		if err != nil {
			return
		}
		test1(minute)
	case "2":
		taskArg := os.Args[2]
		taskNum, err := strconv.Atoi(taskArg)
		println("任务数:" + taskArg)
		if err != nil {
			return
		}
		test2(taskNum)
	}

}

// 40 trainer 20 validator 5 share 1 task
func test1(testMinute int) {
	trainerNum := 6

	csp, _ := sw.NewBCCSP()
	peerNumber := 1
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
				ShardCreator:        bootstrapPeers[0].PublicKey,
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
				MaxBlockNumInMemory: 2,
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
					InitWeight: &proto.Envelope[*proto.GlobalWeight]{
						Payload: &proto.GlobalWeight{
							Iteration: 0,
							ModelHash: "QmUN52MabyZmxGLZZtUF8W6MKm2gSvaxUbAh9BCnS1NPyx",
						},
						Signature: nil,
					},
				}})
			}
		}()
	}

	select {
	case <-time.After(time.Duration(testMinute) * time.Minute):
	}
}

// 3 trainer 1 validator 3 share 1-10 task
func test2(taskNum int) {
	trainerNum := 6
	peerNumber := 1

	csp, _ := sw.NewBCCSP()
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
				ShardCreator:        bootstrapPeers[0].PublicKey,
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
				for j := 0; j < taskNum; j++ {
					peer.CreateTask(&task.Task{TaskGenesis: &proto.TaskGenesis{
						TaskId: fmt.Sprintf("task%d", j),
						InitWeight: &proto.Envelope[*proto.GlobalWeight]{
							Payload: &proto.GlobalWeight{
								Iteration: 0,
								ModelHash: "QmUN52MabyZmxGLZZtUF8W6MKm2gSvaxUbAh9BCnS1NPyx",
							},
							Signature: nil,
						},
					}})
				}
			}
		}()
	}

	select {
	case <-time.After(time.Second * 180):
	}
}
