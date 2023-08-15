package notary

import (
	"encoding/json"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/comm/crypto"
	"github.com/wisecoach/ml_chain/proto"
	"time"
)

func NewNotarySignReqMessageListener(service crypto.MessageCryptoService, node MessageSender, self *proto.RemotePeer) comm.MessageListener {
	return &notarySignReqMessageListener{
		mcs:  service,
		node: node,
		self: self,
	}
}

func NewNotarySignRespMessageListener(manager NotaryManager) comm.MessageListener {
	return &notarySignRespMessageListener{notaryManager: manager}
}

type notarySignReqMessageListener struct {
	mcs  crypto.MessageCryptoService
	node MessageSender
	self *proto.RemotePeer
}

func (n *notarySignReqMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	req := message.Envelope.Payload.GetNotarySignReq()
	payload := req.Transaction.Payload
	bytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	sign, err := n.mcs.Sign(bytes)
	if err != nil {
		return
	}
	msg := &proto.Message{
		Content: &proto.NotarySignRespMessage{
			ChainId:   req.Transaction.Header.ChainId,
			TxId:      req.Transaction.Header.TxId,
			Pk:        n.self.PublicKey,
			Signature: sign,
		},
		Header: &proto.Header{
			Creator:   n.self.PublicKey,
			ChainId:   req.Transaction.Header.ChainId,
			TxId:      "",
			Timestamp: time.Time{},
		},
	}
	n.node.SendToPeers(msg, message.Sender)
}

type notarySignRespMessageListener struct {
	notaryManager NotaryManager
}

func (n *notarySignRespMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	resp := message.Envelope.Payload.GetNotarySignResp()
	n.notaryManager.HandleSignResp(resp)
}
