package ps2messages

import (
	"maps"

	"github.com/x0k/ps2-spy/internal/lib/publisher"
)

type serviceStateChangedHandler chan<- ServiceStateChanged

func (h serviceStateChangedHandler) Type() string {
	return ServiceStateChangedType
}

func (h serviceStateChangedHandler) Handle(msg publisher.Event) {
	h <- *(msg.(*ServiceStateChanged))
}

type heartbeatHandler chan<- Heartbeat

func (h heartbeatHandler) Type() string {
	return HeartbeatType
}

func (h heartbeatHandler) Handle(msg publisher.Event) {
	h <- *(msg.(*Heartbeat))
}

type serviceMessageHandler chan<- ServiceMessage[map[string]any]

func (h serviceMessageHandler) Type() string {
	return ServiceMessageType
}

func (h serviceMessageHandler) Handle(msg publisher.Event) {
	t := msg.(*ServiceMessage[map[string]any])
	h <- ServiceMessage[map[string]any]{
		MessageBase: t.MessageBase,
		Payload:     maps.Clone(t.Payload),
	}
}

type subscriptionSettingsHandler chan<- SubscriptionSettings

func (h subscriptionSettingsHandler) Type() string {
	return SubscriptionSignatureField
}

func (h subscriptionSettingsHandler) Handle(msg publisher.Event) {
	h <- *(msg.(*SubscriptionSettings))
}
