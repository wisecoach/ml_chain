package txhandler

import (
	"bytes"
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
	MaxBlockNumInMemory := 1

	if bytes.Equal(t.node.Self().PublicKey, t.config.ShardCreator) {
		csp, _ := sw.NewBCCSP()
		bootstrapPeers := make([]*proto.RemotePeer, 0)
		privateKeys := make([]bccsp.Key, 0)
		id, _ := strconv.ParseInt(taskGenesis.TaskId[4:], 10, 32)
		basePort := 10000 + int(id*100)
		trainerNumber := t.config.TrainerNum

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
			port := basePort + i
			bootstrapPeers = append(bootstrapPeers, &proto.RemotePeer{
				Endpoint:  "127.0.0.1:" + strconv.Itoa(port),
				PublicKey: pkBytes,
			})
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
			apiPort := strconv.Itoa(6006)
			apiBaseUrl := "http://127.0.0.1:" + apiPort
			config := &trainer.Config{
				TaskId:              taskGenesis.TaskId,
				Cindex:              strconv.Itoa(i),
				Aggregator:          bootstrapPeers[0].PublicKey,
				TrainerNum:          t.config.TrainerNum,
				GenesisBlock:        genesis,
				ApiBaseUrl:          apiBaseUrl,
				Sk:                  t.config.Sk,
				KeyImportOpts:       &bccsp.ECDSAPKIXPublicKeyImportOpts{Temporary: false},
				HashOpts:            &bccsp.SHA256Opts{},
				SignerOpts:          nil,
				Self:                bootstrapPeers[i],
				BootstrapPeers:      bootstrapPeers,
				Notaries:            make([]*proto.RemotePeer, 0),
				TimeoutRPC:          time.Second * 1000,
				MaxBlockNumInMemory: MaxBlockNumInMemory,
			}
			go trainer.New(config)
			t.logger.Info(fmt.Sprintf("bootstrap trainer peer: %s", bootstrapPeers[i].Endpoint))
		}
	}
}

func (t *TaskDeployManager) TxType() reflect.Type {
	return reflect.TypeOf(&proto.TaskGenesis{})
}
