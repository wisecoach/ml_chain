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
	testArg := os.Args[1]
	switch testArg {
	case "1":
		test1()
	case "2":
		test2()
	case "3":
		test3()
	case "4":
		taskArg := os.Args[2]
		taskNum, err := strconv.Atoi(taskArg)
		println("任务数:" + taskArg)
		if err != nil {
			return
		}
		test4(taskNum)
	}

}

// 40 trainer 20 validator 5 share 1 task
func test1() {
	trainerNum := 5
	validatorNum := 1
	NumSharePy := 5
	configFile, err := os.Open("data/init_global_weight.json")
	if err != nil {
		return
	}
	jsonParser := json.NewDecoder(configFile)
	var initWeight []float32
	if err = jsonParser.Decode(&initWeight); err != nil {
	}

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
				ValidatorNum:        validatorNum,
				NumSharePy:          NumSharePy,
				TaskMgrNum:          5,
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
	case <-time.After(time.Second * 1800):
	}
}

// 20 trainer 10 validator 5 share 1 task
func test2() {
	trainerNum := 10
	validatorNum := 5
	NumSharePy := 5
	configFile, err := os.Open("data/init_global_weight.json")
	if err != nil {
		return
	}
	jsonParser := json.NewDecoder(configFile)
	var initWeight []float32
	if err = jsonParser.Decode(&initWeight); err != nil {
	}

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
				ValidatorNum:        validatorNum,
				NumSharePy:          NumSharePy,
				TaskMgrNum:          5,
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
	case <-time.After(time.Second * 1800):
	}
}

// 10 trainer 5 validator 5 share 1 task
func test3() {
	trainerNum := 20
	validatorNum := 10
	NumSharePy := 5
	configFile, err := os.Open("data/init_global_weight.json")
	if err != nil {
		return
	}
	jsonParser := json.NewDecoder(configFile)
	var initWeight []float32
	if err = jsonParser.Decode(&initWeight); err != nil {
	}

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
				ValidatorNum:        validatorNum,
				NumSharePy:          NumSharePy,
				TaskMgrNum:          5,
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
	case <-time.After(time.Second * 1800):
	}
}

// 3 trainer 1 validator 3 share 1-10 task
func test4(taskNum int) {
	trainerNum := 3
	validatorNum := 1
	peerNumber := 1
	configFile, err := os.Open("data/init_global_weight.json")
	if err != nil {
		return
	}
	jsonParser := json.NewDecoder(configFile)
	var initWeight []float32
	if err = jsonParser.Decode(&initWeight); err != nil {
	}

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
				ValidatorNum:        validatorNum,
				NumSharePy:          5,
				TaskMgrNum:          1,
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
			}
		}()
	}

	select {
	case <-time.After(time.Second * 180):
	}
}
