package events_module

import (
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type ServiceMessagePayloadPublisher struct {
	pubsub.Publisher[map[string]any]
}

func newServiceMessagePayloadPublisher(publisher pubsub.Publisher[map[string]any]) *ServiceMessagePayloadPublisher {
	return &ServiceMessagePayloadPublisher{
		Publisher: publisher,
	}
}

func (m *ServiceMessagePayloadPublisher) Publish(msg streaming.Message) {
	if serviceMsg, ok := msg.(streaming.ServiceMessage[map[string]any]); ok {
		m.Publisher.Publish(serviceMsg.Payload)
	}
}
