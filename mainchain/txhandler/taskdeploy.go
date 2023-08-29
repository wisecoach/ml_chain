package txhandler

import (
	"encoding/json"
	"fmt"
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/bccsp/sw"
	"github.com/wisecoach/ml_chain/block/manager"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
	"os/exec"
	"reflect"
	"strconv"
	"time"
)

type TaskDeployManager struct {
	node *node.Node
	mcs  crypto.MessageCryptoService

	config *Config
	logger *zap.Logger
}

func NewTaskDeployManager(config *Config, node2 *node.Node, service crypto.MessageCryptoService) manager.TxHandler {
	return &TaskDeployManager{
		node:   node2,
		mcs:    service,
		config: config,
		logger: log.GetLogger(node2.Self().Endpoint),
	}
}

func (t *TaskDeployManager) HandleTx(tx *proto.Transaction) {
	t.logger.Debug("begin to handle task deployment")
	taskGenesis := tx.Payload.(*proto.TaskGenesis)
	index := -1
	for i, mgrPk := range taskGenesis.ManagerList {
		if reflect.DeepEqual(t.node.Self().PublicKey, mgrPk) {
			index = i
		}
	}
	MaxBlockNumInMemory := 2
	// if self is a manager for the task, bootstrap trainer with self sk
	if index != -1 {
		bootstrapPeers := make([]*proto.RemotePeer, 0)
		transactions := make([]*proto.Envelope[*proto.Transaction], 0)
		transactions = append(transactions, &proto.Envelope[*proto.Transaction]{
			Payload:   tx,
			Signature: nil,
		})

		blockData := &proto.BlockData{
			Transactions: transactions,
		}
		marshal, _ := json.Marshal(blockData)
		dataHash, _ := t.mcs.Hash(marshal)

		genesis := &proto.Block{
			Header: &proto.BlockHeader{
				DataHash:    dataHash,
				PrevHash:    dataHash,
				BlockNumber: 0,
				Miner:       nil,
			},
			Data: blockData,
		}

		id, _ := strconv.ParseInt(taskGenesis.TaskId[4:], 10, 32)
		basePort := 10000 + int(id*100)
		port := basePort + index
		pk, _ := t.config.Sk.PublicKey()
		pkBytes, _ := pk.Bytes()
		self := &proto.RemotePeer{
			Endpoint:  "127.0.0.1:" + strconv.Itoa(port),
			PublicKey: pkBytes,
		}

		for i, pkBytes := range taskGenesis.ManagerList {
			bootstrapPeers = append(bootstrapPeers, &proto.RemotePeer{
				Endpoint:  "127.0.0.1:" + strconv.Itoa(basePort+i),
				PublicKey: pkBytes,
			})
		}
		apiPort := strconv.Itoa(basePort + 50 + index/t.config.NumSharePy)
		apiBaseUrl := "http://localhost:" + apiPort
		if index%t.config.NumSharePy == 0 {
			cmd := exec.Command("./ML/venv/bin/python", "./ML/main.py", apiPort)
			err := cmd.Start()
			if err != nil {
				return
			}
		}
		taskConfig := &trainer.Config{
			TaskId:              taskGenesis.TaskId,
			TrainerNum:          t.config.TrainerNum,
			ValidatorNum:        t.config.ValidatorNum,
			GenesisBlock:        genesis,
			ApiBaseUrl:          apiBaseUrl,
			Sk:                  t.config.Sk,
			KeyImportOpts:       &bccsp.ECDSAPKIXPublicKeyImportOpts{Temporary: false},
			HashOpts:            &bccsp.SHA256Opts{},
			SignerOpts:          nil,
			Self:                self,
			BootstrapPeers:      bootstrapPeers,
			Notaries:            make([]*proto.RemotePeer, 0),
			TimeoutRPC:          time.Second * 1000,
			MaxBlockNumInMemory: MaxBlockNumInMemory,
		}
		go trainer.New(taskConfig)

		t.logger.Info(fmt.Sprintf("bootstrap manager peer: %s", self.Endpoint))
	}

	// if self is the first manager, need to bootstrap trainer for the task, it's just for test, in product environment,
	// trainer will bootstrap by other user
	// otherwise, it also need to bootstrap python server and init task
	if index == 0 {
		csp, _ := sw.NewBCCSP()
		bootstrapPeers := make([]*proto.RemotePeer, 0)
		privateKeys := make([]bccsp.Key, 0)
		id, _ := strconv.ParseInt(taskGenesis.TaskId[4:], 10, 32)
		basePort := 10000 + int(id*100)
		mgrNum := len(taskGenesis.ManagerList)
		trainerNumber := t.config.TrainerNum - mgrNum
		// add all manager
		for i, pkBytes := range taskGenesis.ManagerList {
			bootstrapPeers = append(bootstrapPeers, &proto.RemotePeer{
				Endpoint:  "127.0.0.1:" + strconv.Itoa(basePort+i),
				PublicKey: pkBytes,
			})
		}

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
			port := basePort + i + mgrNum
			apiPort := strconv.Itoa(basePort + 50 + (i+mgrNum)/t.config.NumSharePy)
			bootstrapPeers = append(bootstrapPeers, &proto.RemotePeer{
				Endpoint:  "127.0.0.1:" + strconv.Itoa(port),
				PublicKey: pkBytes,
			})
			if (i+mgrNum)%t.config.NumSharePy == 0 {
				cmd := exec.Command("./ML/venv/bin/python", "./ML/main.py", apiPort)
				err := cmd.Start()
				if err != nil {
					return
				}
			}
		}

		transactions := make([]*proto.Envelope[*proto.Transaction], 0)
		transactions = append(transactions, &proto.Envelope[*proto.Transaction]{
			Payload:   tx,
			Signature: nil,
		})

		blockData := &proto.BlockData{
			Transactions: transactions,
		}
		marshal, _ := json.Marshal(blockData)
		dataHash, _ := t.mcs.Hash(marshal)

		genesis := &proto.Block{
			Header: &proto.BlockHeader{
				DataHash:    dataHash,
				PrevHash:    dataHash,
				BlockNumber: 0,
				Miner:       nil,
			},
			Data: blockData,
		}

		for i := 0; i < trainerNumber; i++ {
			apiPort := strconv.Itoa(basePort + 50 + (i+mgrNum)/t.config.NumSharePy)
			apiBaseUrl := "http://127.0.0.1:" + apiPort
			config := &trainer.Config{
				TaskId:              taskGenesis.TaskId,
				TrainerNum:          t.config.TrainerNum,
				ValidatorNum:        t.config.ValidatorNum,
				GenesisBlock:        genesis,
				ApiBaseUrl:          apiBaseUrl,
				Sk:                  t.config.Sk,
				KeyImportOpts:       &bccsp.ECDSAPKIXPublicKeyImportOpts{Temporary: false},
				HashOpts:            &bccsp.SHA256Opts{},
				SignerOpts:          nil,
				Self:                bootstrapPeers[i+mgrNum],
				BootstrapPeers:      bootstrapPeers,
				Notaries:            make([]*proto.RemotePeer, 0),
				TimeoutRPC:          time.Second * 1000,
				MaxBlockNumInMemory: MaxBlockNumInMemory,
			}
			go trainer.New(config)
			t.logger.Info(fmt.Sprintf("bootstrap trainer peer: %s", bootstrapPeers[i+mgrNum].Endpoint))

		}
	}
}

func (t *TaskDeployManager) TxType() reflect.Type {
	return reflect.TypeOf(&proto.TaskGenesis{})
}
