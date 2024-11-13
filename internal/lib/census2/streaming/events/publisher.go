package events

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

var ErrInvalidEvent = fmt.Errorf("invalid event")
var ErrEventNameNotFound = fmt.Errorf("event name not found")
var ErrUnknownEventName = fmt.Errorf("unknown event name")

type bufferedPublisher struct {
	pubsub.Publisher[EventType]
	buffers map[EventType]Event
}

func NewPublisher(publisher pubsub.Publisher[EventType]) *bufferedPublisher {
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

func (p *bufferedPublisher) Publish(event streaming.Event) error {
	msg, ok := event.(streaming.MessageReceived)
	if !ok {
		return ErrInvalidEvent
	}
	name, ok := msg[core.EventNameField].(string)
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
