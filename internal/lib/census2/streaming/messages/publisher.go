package ps2messages

import (
	"fmt"
	"log/slog"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

var ErrUnknownMessageType = fmt.Errorf("unknown message type")

type Publisher struct {
	log                      *slog.Logger
	msgBuff                  core.MessageBase
	serviceStateChangedBuff  ServiceStateChanged
	ServiceStateChanged      chan ServiceStateChanged
	heartbeatBuff            Heartbeat
	Heartbeat                chan Heartbeat
	serviceMessageBuff       ServiceMessage[map[string]any]
	ServiceMessage           chan ServiceMessage[map[string]any]
	subscriptionSettingsBuff SubscriptionSettings
	SubscriptionSettings     chan SubscriptionSettings
}

func NewPublisher(log *slog.Logger) *Publisher {
	return &Publisher{
		log: log.With(
			slog.String("component", "census2.streaming.messages.Publisher"),
		),
		ServiceStateChanged:  make(chan ServiceStateChanged),
		Heartbeat:            make(chan Heartbeat),
		ServiceMessage:       make(chan ServiceMessage[map[string]any]),
		SubscriptionSettings: make(chan SubscriptionSettings),
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
		err = mapstructure.Decode(s, &p.subscriptionSettingsBuff)
		if err != nil {
			return
		}
		p.SubscriptionSettings <- p.subscriptionSettingsBuff
	}
	// Ignore help message
	if _, ok := msg[HelpSignatureField]; ok {
		return
	}
	err = core.AsMessageBase(msg, &p.msgBuff)
	if err != nil {
		return
	}
	if p.msgBuff.Service != core.EventService {
		p.log.Warn("non event message is dropped", slog.Any("msg", msg))
		return
	}
	switch p.msgBuff.Type {
	case ServiceMessageType:
		err = mapstructure.Decode(msg, &p.serviceMessageBuff)
		if err != nil {
			return
		}
		p.ServiceMessage <- p.serviceMessageBuff
	case HeartbeatType:
		err = mapstructure.Decode(msg, &p.heartbeatBuff)
		if err != nil {
			return
		}
		p.Heartbeat <- p.heartbeatBuff
	case ServiceStateChangedType:
		err = mapstructure.Decode(msg, &p.serviceStateChangedBuff)
		if err != nil {
			return
		}
		p.ServiceStateChanged <- p.serviceStateChangedBuff
	default:
		err = fmt.Errorf("%s: %w", p.msgBuff.Type, ErrUnknownMessageType)
	}
}
