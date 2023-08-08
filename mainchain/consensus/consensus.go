package consensus

import "github.com/wisecoach/ml_chain/proto"

type MainChainConsensus struct {
}

func (m *MainChainConsensus) Order(transaction *proto.Envelope[*proto.Transaction]) {
	// TODO implement me
	panic("implement me")
}

func (m *MainChainConsensus) Consensus(block *proto.Envelope[*proto.Block]) {
	// TODO implement me
	panic("implement me")
}

func (m *MainChainConsensus) Start() {
	// TODO implement me
	panic("implement me")
}
