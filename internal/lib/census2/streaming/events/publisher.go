package events

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

var ErrInvalidEvent = fmt.Errorf("invalid event")
var ErrEventNameNotFound = fmt.Errorf("event name not found")
var ErrUnknownEventName = fmt.Errorf("unknown event name")

type eventsPublisher struct {
	pubsub.Publisher[Event]
	buffers map[EventType]Event
	onError func(err error)
}

func NewPublisher(
	publisher pubsub.Publisher[Event],
	onError func(err error),
) *eventsPublisher {
	return &eventsPublisher{
		Publisher: publisher,
		buffers: map[EventType]Event{
			AchievementEarnedEventName:     AchievementEarned{},
			BattleRankUpEventName:          BattleRankUp{},
			DeathEventName:                 Death{},
			GainExperienceEventName:        GainExperience{},
			ItemAddedEventName:             ItemAdded{},
			PlayerFacilityCaptureEventName: PlayerFacilityCapture{},
			PlayerFacilityDefendEventName:  PlayerFacilityDefend{},
			PlayerLoginEventName:           PlayerLogin{},
			PlayerLogoutEventName:          PlayerLogout{},
			SkillAddedEventName:            SkillAdded{},
			VehicleDestroyEventName:        VehicleDestroy{},
			ContinentLockEventName:         ContinentLock{},
			FacilityControlEventName:       FacilityControl{},
			MetagameEventEventName:         MetagameEvent{},
		},
		onError: onError,
	}
}

func (p *eventsPublisher) Publish(event map[string]any) {
	name, ok := event[core.EventNameField].(string)
	if !ok {
		p.onError(ErrEventNameNotFound)
		return
	}
	buff, ok := p.buffers[EventType(name)]
	if !ok {
		p.onError(fmt.Errorf("%s: %w", name, ErrUnknownEventName))
		return
	}
	if err := mapstructure.Decode(event, &buff); err != nil {
		p.onError(fmt.Errorf("%q decoding: %w", name, err))
		return
	}
	p.Publisher.Publish(buff)
}
