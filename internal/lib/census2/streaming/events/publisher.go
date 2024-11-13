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

type bufferedPublisher struct {
	pubsub.Publisher[Event]
	buffers map[EventType]Event
}

func NewPublisher(publisher pubsub.Publisher[Event]) *bufferedPublisher {
	return &bufferedPublisher{
		Publisher: publisher,
		buffers: map[EventType]Event{
			AchievementEarnedEventName:     &AchievementEarned{},
			BattleRankUpEventName:          &BattleRankUp{},
			DeathEventName:                 &Death{},
			GainExperienceEventName:        &GainExperience{},
			ItemAddedEventName:             &ItemAdded{},
			PlayerFacilityCaptureEventName: &PlayerFacilityCapture{},
			PlayerFacilityDefendEventName:  &PlayerFacilityDefend{},
			PlayerLoginEventName:           &PlayerLogin{},
			PlayerLogoutEventName:          &PlayerLogout{},
			SkillAddedEventName:            &SkillAdded{},
			VehicleDestroyEventName:        &VehicleDestroy{},
			ContinentLockEventName:         &ContinentLock{},
			FacilityControlEventName:       &FacilityControl{},
			MetagameEventEventName:         &MetagameEvent{},
		},
	}
}

func (p *bufferedPublisher) Publish(event map[string]any) error {
	name, ok := event[core.EventNameField].(string)
	if !ok {
		return ErrEventNameNotFound
	}
	if buff, ok := p.buffers[EventType(name)]; ok {
		err := mapstructure.Decode(event, buff)
		if err != nil {
			return fmt.Errorf("%q decoding: %w", name, err)
		}
		return p.Publisher.Publish(buff)
	}
	return fmt.Errorf("%s: %w", name, ErrUnknownEventName)
}
