package ps2messages

import "maps"

type messageHandler interface {
	Type() string
	Handle(msg any)
}

type serviceStateChangedHandler chan<- ServiceStateChanged

func (h serviceStateChangedHandler) Type() string {
	return ServiceStateChangedType
}

func (h serviceStateChangedHandler) Handle(msg any) {
	h <- *(msg.(*ServiceStateChanged))
}

type heartbeatHandler chan<- Heartbeat

func (h heartbeatHandler) Type() string {
	return HeartbeatType
}

func (h heartbeatHandler) Handle(msg any) {
	h <- *(msg.(*Heartbeat))
}

type serviceMessageHandler chan<- ServiceMessage[map[string]any]

func (h serviceMessageHandler) Type() string {
	return ServiceMessageType
}

func (h serviceMessageHandler) Handle(msg any) {
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

func (h subscriptionSettingsHandler) Handle(msg any) {
	h <- *(msg.(*SubscriptionSettings))
}

func handlerForInterface(handler any) messageHandler {
	switch v := handler.(type) {
	case chan ServiceStateChanged:
		return serviceStateChangedHandler(v)
	case chan<- ServiceStateChanged:
		return serviceStateChangedHandler(v)
	case chan Heartbeat:
		return heartbeatHandler(v)
	case chan<- Heartbeat:
		return heartbeatHandler(v)
	case chan ServiceMessage[map[string]any]:
		return serviceMessageHandler(v)
	case chan<- ServiceMessage[map[string]any]:
		return serviceMessageHandler(v)
	case chan SubscriptionSettings:
		return subscriptionSettingsHandler(v)
	case chan<- SubscriptionSettings:
		return subscriptionSettingsHandler(v)
	}
	return nil
}
