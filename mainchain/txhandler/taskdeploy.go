package txhandler

import (
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
	// if self is a manager for the task, bootstrap trainer with self sk
	if index != -1 {
		bootstrapPeers := make([]*proto.RemotePeer, 0)
		transactions := make([]*proto.Envelope[*proto.Transaction], 0)
		transactions = append(transactions, &proto.Envelope[*proto.Transaction]{
			Payload:   tx,
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

		taskConfig := &trainer.Config{
			TaskId:         taskGenesis.TaskId,
			ValidatorNum:   1,
			GenesisBlock:   genesis,
			ApiBaseUrl:     "http://localhost:8999",
			Sk:             t.config.Sk,
			KeyImportOpts:  &bccsp.ECDSAPKIXPublicKeyImportOpts{Temporary: false},
			HashOpts:       &bccsp.SHA256Opts{},
			SignerOpts:     nil,
			Self:           self,
			BootstrapPeers: bootstrapPeers,
			TimeoutRPC:     time.Second * 1000,
		}
		go trainer.New(taskConfig)

		t.logger.Info(fmt.Sprintf("bootstrap manager peer: %s", self.Endpoint))
	}

	// if self is the first manager, need to bootstrap trainer for the task, it's just for test, in product environment, trainer will bootstrap by other user
	if index == 0 {
		csp, _ := sw.NewBCCSP()
		trainerNumber := t.config.TrainerNum
		bootstrapPeers := make([]*proto.RemotePeer, 0)
		privateKeys := make([]bccsp.Key, 0)
		id, _ := strconv.ParseInt(taskGenesis.TaskId[4:], 10, 32)
		basePort := 10000 + int(id*100)
		mgrNum := len(taskGenesis.ManagerList)
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
			bootstrapPeers = append(bootstrapPeers, &proto.RemotePeer{
				Endpoint:  "127.0.0.1:" + strconv.Itoa(basePort+i+mgrNum),
				PublicKey: pkBytes,
			})
		}

		transactions := make([]*proto.Envelope[*proto.Transaction], 0)
		transactions = append(transactions, &proto.Envelope[*proto.Transaction]{
			Payload:   tx,
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
				TaskId:         taskGenesis.TaskId,
				ValidatorNum:   1,
				GenesisBlock:   genesis,
				ApiBaseUrl:     "http://localhost:8999",
				Sk:             t.config.Sk,
				KeyImportOpts:  &bccsp.ECDSAPKIXPublicKeyImportOpts{Temporary: false},
				HashOpts:       &bccsp.SHA256Opts{},
				SignerOpts:     nil,
				Self:           bootstrapPeers[i+mgrNum],
				BootstrapPeers: bootstrapPeers,
				TimeoutRPC:     time.Second * 1000,
			}
			go trainer.New(config)
			t.logger.Info(fmt.Sprintf("bootstrap trainer peer: %s", bootstrapPeers[i+mgrNum].Endpoint))
		}
	}
}

func (t *TaskDeployManager) TxType() reflect.Type {
	return reflect.TypeOf(&proto.TaskGenesis{})
}
