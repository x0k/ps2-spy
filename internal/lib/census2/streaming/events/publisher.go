package events

import (
	"encoding/json"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

var ErrInvalidEvent = fmt.Errorf("invalid event")
var ErrEventNameNotFound = fmt.Errorf("event name not found")
var ErrUnknownEventName = fmt.Errorf("unknown event name")

type eventsPublisher struct {
	publisher pubsub.Publisher[Event]
	onError   func(err error)
}

func NewPublisher(
	publisher pubsub.Publisher[Event],
	onError func(err error),
) *eventsPublisher {
	return &eventsPublisher{
		publisher: publisher,
		onError:   onError,
	}
}

func (p *eventsPublisher) Publish(event json.RawMessage) {
	var base core.EventBase
	err := json.Unmarshal(event, &base)
	if err != nil {
		p.onError(ErrInvalidEvent)
	}
	switch EventType(base.EventName) {
	case AchievementEarnedEventName:
		parseAndPublish[AchievementEarned](p, event)
	case BattleRankUpEventName:
		parseAndPublish[BattleRankUp](p, event)
	case DeathEventName:
		parseAndPublish[Death](p, event)
	case GainExperienceEventName:
		parseAndPublish[GainExperience](p, event)
	case ItemAddedEventName:
		parseAndPublish[ItemAdded](p, event)
	case PlayerFacilityCaptureEventName:
		parseAndPublish[PlayerFacilityCapture](p, event)
	case PlayerFacilityDefendEventName:
		parseAndPublish[PlayerFacilityDefend](p, event)
	case PlayerLoginEventName:
		parseAndPublish[PlayerLogin](p, event)
	case PlayerLogoutEventName:
		parseAndPublish[PlayerLogout](p, event)
	case SkillAddedEventName:
		parseAndPublish[SkillAdded](p, event)
	case VehicleDestroyEventName:
		parseAndPublish[VehicleDestroy](p, event)
	case ContinentLockEventName:
		parseAndPublish[ContinentLock](p, event)
	case FacilityControlEventName:
		parseAndPublish[FacilityControl](p, event)
	case MetagameEventEventName:
		parseAndPublish[MetagameEvent](p, event)
	default:
		p.onError(fmt.Errorf("%w: %s", ErrUnknownEventName, base.EventName))
	}
}

func parseAndPublish[T Event](p *eventsPublisher, rawMsg json.RawMessage) {
	var msg T
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		p.onError(fmt.Errorf("failed to decode service message: %w", err))
		return
	}
	p.publisher.Publish(msg)
}
