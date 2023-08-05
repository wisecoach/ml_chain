package message

import (
	"github.com/wisecoach/ml_chain/comm/node"
	"github.com/wisecoach/ml_chain/proto"
	"github.com/wisecoach/ml_chain/trainer/model/aggregate"
	"github.com/wisecoach/ml_chain/util/log"
	"go.uber.org/zap"
)

type LocalWightMessageListener struct {
	aggregator aggregate.Aggregator
	logger     *zap.Logger
}

func NewLocalWightMessageListener(aggregator aggregate.Aggregator) node.MessageListener {
	return &LocalWightMessageListener{
		aggregator: aggregator,
		logger:     log.GetLogger(),
	}
}

func (l *LocalWightMessageListener) HandleMessage(message *proto.ReceivedMessage) {
	weight := message.Envelope.Payload.GetSubmitLocalityWeight()
	err := l.aggregator.HandleLocalModel(weight.LocalityWeight)
	if err != nil {
		l.logger.Error("handle submitted local model failed with err: " + err.Error())
		return
	}
}
