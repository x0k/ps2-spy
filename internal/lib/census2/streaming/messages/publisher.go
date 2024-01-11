package ps2messages

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

var ErrUnknownMessageType = fmt.Errorf("unknown message type")
var ErrUnknownMessageHandler = fmt.Errorf("unknown message handler")

type Publisher struct {
	log                      *slog.Logger
	msgBaseBuff              core.MessageBase
	handlersMu               sync.RWMutex
	handlers                 map[string][]messageHandler
	subscriptionSettingsBuff *SubscriptionSettings
	buffers                  map[string]any
}

func NewPublisher(log *slog.Logger) *Publisher {
	return &Publisher{
		log: log.With(
			slog.String("component", "census2.streaming.messages.Publisher"),
		),
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

func (p *Publisher) Publish(msg map[string]any) {
	var err error
	defer func() {
		if err != nil {
			p.log.Warn("failed to publish message", slog.Any("msg", msg), sl.Err(err))
		}
	}()
	// Subscription
	if s, ok := msg[SubscriptionSignatureField]; ok {
		err = mapstructure.Decode(s, p.subscriptionSettingsBuff)
		if err != nil {
			return
		}
		p.publish(SubscriptionSignatureField, p.subscriptionSettingsBuff)
		return
	}
	// Ignore help message
	if _, ok := msg[HelpSignatureField]; ok {
		return
	}
	err = core.AsMessageBase(msg, &p.msgBaseBuff)
	if err != nil {
		return
	}
	if p.msgBaseBuff.Service != core.EventService {
		p.log.Warn("non event message is dropped", slog.Any("msg", msg))
		return
	}
	if buff, ok := p.buffers[p.msgBaseBuff.Type]; ok {
		err = mapstructure.Decode(msg, buff)
		if err != nil {
			return
		}
		p.publish(p.msgBaseBuff.Type, buff)
		return
	}
	err = fmt.Errorf("%s: %w", p.msgBaseBuff.Type, ErrUnknownMessageType)
}
