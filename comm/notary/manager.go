package notary

import "github.com/wisecoach/ml_chain/proto"

type NotaryManager interface {
	//
	// GetNotaries
	//  @Description: get all the notaries of chainId
	//  @param chainId
	//  @return *proto.RemotePeer
	//  @return error when chainId is not register or cannot find remotePeer for pk
	//
	GetNotaries(chainId string) ([]*proto.RemotePeer, error)

	//
	// RegisterNotary
	//  @Description: register notary to specific chain, it should be called when handle transaction
	//
	RegisterNotary(chainId string, pk []byte)

	//
	// Discover
	//  @Description: discover notary
	//
	Discover(notary *proto.RemotePeer)

	//
	// SignCrossTx
	//  @Description: sign the cross transaction asynchronously
	//  @param transaction to sign
	//  @return chan for signed transaction
	//
	SignCrossTx(transaction *proto.Transaction) (chan *proto.Transaction, error)

	//
	// HandleSignResp
	//  @Description: handle the NotarySignRespMessage
	//  @param resp
	//
	HandleSignResp(resp *proto.NotarySignRespMessage)
}
