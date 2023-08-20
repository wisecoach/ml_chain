package message

import (
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/aggregate"
)

type LocalWightMessageListener struct {
	aggregator aggregate.Aggregator
}

func NewLocalWightMessageListener(aggregator aggregate.Aggregator) comm.MessageListener {
	return &LocalWightMessageListener{
		aggregator: aggregator,
	}
}

func (l *LocalWightMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	weight := message.Envelope.Payload.GetSubmitLocalityWeight()
	l.aggregator.HandleLocalModel(weight.LocalityWeight)
}
