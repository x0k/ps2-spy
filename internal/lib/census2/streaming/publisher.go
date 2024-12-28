package streaming

import (
	"encoding/json"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

var ErrUnknownEventType = fmt.Errorf("unknown event type")
var ErrUnknownMessageType = fmt.Errorf("unknown message type")
var ErrUnknownMessageHandler = fmt.Errorf("unknown message handler")
var ErrUnsupportedMessageService = fmt.Errorf("unsupported message service")

type Publisher struct {
	publisher pubsub.Publisher[Message]
	onError   func(err error)
}

func NewPublisher(
	publisher pubsub.Publisher[Message],
	onError func(err error),
) *Publisher {
	return &Publisher{
		publisher: publisher,
		onError:   onError,
	}
}

func (p *Publisher) Publish(msg json.RawMessage) {
	var content map[string]json.RawMessage
	if err := json.Unmarshal(msg, &content); err != nil {
		p.onError(fmt.Errorf("failed to decode message base: %w", err))
		return
	}
	// Ignore help message
	if _, ok := content[HelpSignatureField]; ok {
		return
	}
	// Subscription
	if s, ok := content[SubscriptionSignatureField]; ok {
		var subscriptionSettings SubscriptionSettings
		err := json.Unmarshal(s, &subscriptionSettings)
		if err != nil {
			p.onError(fmt.Errorf("%q decoding: %w", SubscriptionSignatureField, err))
			return
		}
		p.publisher.Publish(subscriptionSettings)
		return
	}
	var base core.MessageBase
	err := json.Unmarshal(msg, &base)
	if err != nil {
		p.onError(fmt.Errorf("failed to decode message base: %w", err))
		return
	}
	if base.Service != core.EventService {
		p.onError(fmt.Errorf("%w: %s", ErrUnsupportedMessageService, base.Service))
		return
	}
	switch MessageType(base.Type) {
	case ServiceStateChangedType:
		parseAndPublish[ServiceStateChanged](p, msg)
	case HeartbeatType:
		parseAndPublish[Heartbeat](p, msg)
	case ServiceMessageType:
		parseAndPublish[ServiceMessage[json.RawMessage]](p, msg)
	default:
		p.onError(fmt.Errorf("%w: %s", ErrUnknownMessageType, base.Type))
		return
	}
}

func parseAndPublish[T Message](p *Publisher, rawMsg json.RawMessage) {
	var msg T
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		p.onError(fmt.Errorf("failed to decode service message: %w", err))
		return
	}
	p.publisher.Publish(msg)
}
