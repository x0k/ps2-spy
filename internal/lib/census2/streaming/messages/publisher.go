package ps2messages

import (
	"fmt"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
)

var ErrUnknownMessageType = fmt.Errorf("unknown message type")
var ErrUnknownMessageHandler = fmt.Errorf("unknown message handler")
var ErrUnsupportedMessageService = fmt.Errorf("unsupported message service")

type Publisher struct {
	msgBaseBuff              core.MessageBase
	handlersMu               sync.RWMutex
	handlers                 map[string][]messageHandler
	subscriptionSettingsBuff *SubscriptionSettings
	buffers                  map[string]any
}

func NewPublisher() *Publisher {
	return &Publisher{
		handlers:                 map[string][]messageHandler{},
		subscriptionSettingsBuff: &SubscriptionSettings{},
		buffers: map[string]any{
			ServiceStateChangedType: &ServiceStateChanged{},
			HeartbeatType:           &Heartbeat{},
			ServiceMessageType:      &ServiceMessage[map[string]any]{},
		},
	}
}

func (p *Publisher) removeHandler(msgType string, h messageHandler) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	for i, v := range p.handlers[msgType] {
		if v == h {
			p.handlers[msgType] = append(p.handlers[msgType][:i], p.handlers[msgType][i+1:]...)
			return
		}
	}
}

func (p *Publisher) addHandler(h messageHandler) func() {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	p.handlers[h.Type()] = append(p.handlers[h.Type()], h)
	return func() {
		p.removeHandler(h.Type(), h)
	}
}

func (p *Publisher) AddHandler(h any) (func(), error) {
	handler := handlerForInterface(h)
	if handler == nil {
		return nil, ErrUnknownMessageHandler
	}
	return p.addHandler(handler), nil
}

func (p *Publisher) publish(msgType string, msg any) {
	p.handlersMu.RLock()
	defer p.handlersMu.RUnlock()
	for _, h := range p.handlers[msgType] {
		h.Handle(msg)
	}
}

func (p *Publisher) Publish(msg map[string]any) error {
	// Subscription
	if s, ok := msg[SubscriptionSignatureField]; ok {
		err := mapstructure.Decode(s, p.subscriptionSettingsBuff)
		if err != nil {
			return fmt.Errorf("%q decoding: %w", SubscriptionSignatureField, err)
		}
		p.publish(SubscriptionSignatureField, p.subscriptionSettingsBuff)
		return nil
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
		p.publish(p.msgBaseBuff.Type, buff)
		return nil
	}
	return fmt.Errorf("%s: %w", p.msgBaseBuff.Type, ErrUnknownMessageType)
}
