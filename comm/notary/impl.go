package notary

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
	"sync"
)

type MessageSender interface {
	SendToPeers(msg *proto.Message, peers ...*proto.RemotePeer)
}

func New(msgSender MessageSender) NotaryManager {
	return &notaryManagerImpl{
		lock:          sync.RWMutex{},
		signWaitGroup: sync.WaitGroup{},
		chainToPks:    make(map[string][][]byte),
		pkToNotary:    make(map[string]*proto.RemotePeer),
		txsToSign:     make(map[string]*proto.Transaction),
		msgSender:     msgSender,
	}
}

type notaryManagerImpl struct {
	lock          sync.RWMutex
	signWaitGroup sync.WaitGroup

	chainToPks map[string][][]byte
	pkToNotary map[string]*proto.RemotePeer
	txsToSign  map[string]*proto.Transaction

	msgSender MessageSender
}

func (n *notaryManagerImpl) GetNotaries(chainId string) ([]*proto.RemotePeer, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	notaries := make([]*proto.RemotePeer, 0)
	pks, exists := n.chainToPks[chainId]
	if !exists {
		return nil, errors.New(fmt.Sprintf("chain %s is not registered", chainId))
	}
	for _, pk := range pks {
		pkStr := base64.StdEncoding.EncodeToString(pk)
		notary, exists := n.pkToNotary[pkStr]
		if !exists {
			return nil, errors.New(fmt.Sprintf("notary for pk %s is not discovered", pkStr))
		}
		notaries = append(notaries, notary)
	}
	return notaries, nil
}

func (n *notaryManagerImpl) RegisterNotary(chainId string, pk []byte) {
	n.lock.Lock()
	defer n.lock.Unlock()

	pks, exists := n.chainToPks[chainId]
	if !exists {
		pks = make([][]byte, 0)
		n.chainToPks[chainId] = pks
	}
	pks = append(pks, pk)
}

func (n *notaryManagerImpl) Discover(notary *proto.RemotePeer) {
	n.lock.Lock()
	defer n.lock.Unlock()

	pk := base64.StdEncoding.EncodeToString(notary.PublicKey)
	n.pkToNotary[pk] = notary
}

func (n *notaryManagerImpl) SignCrossTx(transaction *proto.Transaction) (chan *proto.Transaction, error) {
	txChan := make(chan *proto.Transaction)
	chainId := transaction.Header.ChainId
	txId := transaction.Header.TxId
	notaries, err := n.GetNotaries(chainId)
	if err != nil {
		return nil, err
	}
	n.signWaitGroup.Add(len(notaries))
	reqMsg := &proto.Message{
		Content: &proto.NotarySignReqMessage{Transaction: transaction},
		Header:  nil,
	}

	n.lock.Lock()
	n.txsToSign[txId] = transaction
	n.lock.Unlock()

	// send req to notaries
	n.msgSender.SendToPeers(reqMsg, notaries...)
	go func() {
		// wait for enough resp asynchronously
		n.signWaitGroup.Wait()
		// send tx to chan
		n.lock.Lock()
		tx := n.txsToSign[txId]
		txChan <- tx
		delete(n.txsToSign, txId)
		n.lock.Unlock()
	}()

	return txChan, nil
}

func (n *notaryManagerImpl) HandleSignResp(resp *proto.NotarySignRespMessage) {
	n.lock.Lock()
	defer n.lock.Unlock()

	txId := resp.TxId
	pks := n.chainToPks[resp.ChainId]
	// find index to set signature
	index := -1
	for i, pk := range pks {
		if reflect.DeepEqual(pk, resp.Pk) {
			index = i
		}
	}
	tx := n.txsToSign[txId]
	tx.NotarySigns[index] = resp.Signature
	n.signWaitGroup.Done()
}
