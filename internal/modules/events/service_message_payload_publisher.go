package events_module

import (
	"encoding/json"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type ServiceMessagePayloadPublisher struct {
	pubsub.Publisher[json.RawMessage]
}

func newServiceMessagePayloadPublisher(publisher pubsub.Publisher[json.RawMessage]) *ServiceMessagePayloadPublisher {
	return &ServiceMessagePayloadPublisher{
		Publisher: publisher,
	}
}

func (m *ServiceMessagePayloadPublisher) Publish(msg streaming.Message) {
	if serviceMsg, ok := msg.(streaming.ServiceMessage[json.RawMessage]); ok {
		m.Publisher.Publish(serviceMsg.Payload)
	}
}
