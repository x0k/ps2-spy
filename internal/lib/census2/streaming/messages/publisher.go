package ps2messages

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
)

var ErrUnknownMessageType = fmt.Errorf("unknown message type")
var ErrUnknownMessageHandler = fmt.Errorf("unknown message handler")
var ErrUnsupportedMessageService = fmt.Errorf("unsupported message service")

type Publisher struct {
	publisher.Publisher[publisher.Event]
	msgBaseBuff              core.MessageBase
	subscriptionSettingsBuff *SubscriptionSettings
	buffers                  map[string]any
}

func NewPublisher(publisher publisher.Publisher[publisher.Event]) *Publisher {
	return &Publisher{
		Publisher:                publisher,
		subscriptionSettingsBuff: &SubscriptionSettings{},
		buffers: map[string]any{
			ServiceStateChangedType: &ServiceStateChanged{},
			HeartbeatType:           &Heartbeat{},
			ServiceMessageType:      &ServiceMessage[map[string]any]{},
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
		return p.Publisher.Publish(p.subscriptionSettingsBuff)
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
	if buff, ok := p.buffers[p.msgBaseBuff.Type]; ok {
		err = mapstructure.Decode(msg, buff)
		if err != nil {
			return fmt.Errorf("%q decoding: %w", p.msgBaseBuff.Type, err)
		}
		if e, ok := buff.(publisher.Event); ok {
			return p.Publisher.Publish(e)
		}
	}
	return fmt.Errorf("%s: %w", p.msgBaseBuff.Type, ErrUnknownMessageType)
}

func (p *Publisher) AddServiceStateChangedHandler(c chan<- ServiceStateChanged) func() {
	return p.AddHandler(serviceStateChangedHandler(c))
}

func (p *Publisher) AddHeartbeatHandler(c chan<- Heartbeat) func() {
	return p.AddHandler(heartbeatHandler(c))
}

func (p *Publisher) AddServiceMessageHandler(c chan<- ServiceMessage[map[string]any]) func() {
	return p.AddHandler(serviceMessageHandler(c))
}

func (p *Publisher) AddSubscriptionSettingsHandler(c chan<- SubscriptionSettings) func() {
	return p.AddHandler(subscriptionSettingsHandler(c))
}
