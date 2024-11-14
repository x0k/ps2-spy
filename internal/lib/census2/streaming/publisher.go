package streaming

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

var ErrUnknownEventType = fmt.Errorf("unknown event type")
var ErrUnknownMessageType = fmt.Errorf("unknown message type")
var ErrUnknownMessageHandler = fmt.Errorf("unknown message handler")
var ErrUnsupportedMessageService = fmt.Errorf("unsupported message service")

type Publisher struct {
	pubsub.Publisher[Message]
	msgBaseBuff              core.MessageBase
	subscriptionSettingsBuff *SubscriptionSettings
	buffers                  map[MessageType]Message
}

func NewPublisher(publisher pubsub.Publisher[Message]) *Publisher {
	return &Publisher{
		Publisher:                publisher,
		subscriptionSettingsBuff: &SubscriptionSettings{},
		buffers: map[MessageType]Message{
			ServiceStateChangedType: ServiceStateChanged{},
			HeartbeatType:           Heartbeat{},
			ServiceMessageType:      ServiceMessage[map[string]any]{},
		},
	}
}

func (p *Publisher) Publish(msg map[string]any) error {
	// Subscription
	if s, ok := msg[SubscriptionSignatureField]; ok {
		err := mapstructure.Decode(s, p.subscriptionSettingsBuff)
		if err != nil {
			return fmt.Errorf("%q decoding: %w", SubscriptionSignatureField, err)
		}
		return p.Publisher.Publish(*p.subscriptionSettingsBuff)
	}
	// Ignore help message
	if _, ok := msg[HelpSignatureField]; ok {
		return nil
	}
	err := core.AsMessageBase(msg, &p.msgBaseBuff)
	if err != nil {
		return err
	}
	if p.msgBaseBuff.Service != core.EventService {
		return fmt.Errorf("%s: %w", p.msgBaseBuff.Service, ErrUnsupportedMessageService)
	}
	if buff, ok := p.buffers[MessageType(p.msgBaseBuff.Type)]; ok {
		err = mapstructure.Decode(msg, &buff)
		if err != nil {
			return fmt.Errorf("%q decoding: %w", p.msgBaseBuff.Type, err)
		}
		return p.Publisher.Publish(buff)
	}
	return fmt.Errorf("%s: %w", p.msgBaseBuff.Type, ErrUnknownMessageType)
}
