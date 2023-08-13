package message

import (
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/aggregate"
)

type LocalWightMessageListener struct {
	aggregator aggregate.Aggregator
}

func NewLocalWightMessageListener(aggregator aggregate.Aggregator) node.MessageListener {
	return &LocalWightMessageListener{
		aggregator: aggregator,
	}
}

func (l *LocalWightMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	weight := message.Envelope.Payload.GetSubmitLocalityWeight()
	err := l.aggregator.HandleLocalModel(weight.LocalityWeight)
	if err != nil {
		return
	}
}
