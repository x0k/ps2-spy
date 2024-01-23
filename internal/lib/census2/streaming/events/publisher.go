package ps2events

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
)

var ErrEventNameNotFound = fmt.Errorf("event name not found")
var ErrUnknownEventName = fmt.Errorf("unknown event name")

type Publisher struct {
	publisher publisher.Abstract[publisher.Event]
	buffers   map[string]any
}

func NewPublisher(publisher publisher.Abstract[publisher.Event]) *Publisher {
	return &Publisher{
		publisher: publisher,
		buffers: map[string]any{
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

func (p *Publisher) Publish(event map[string]any) error {
	name, ok := event[core.EventNameField].(string)
	if !ok {
		return ErrEventNameNotFound
	}
	if buff, ok := p.buffers[name]; ok {
		err := mapstructure.Decode(event, buff)
		if err != nil {
			return fmt.Errorf("%q decoding: %w", name, err)
		}
		if e, ok := buff.(publisher.Event); ok {
			return p.publisher.Publish(e)
		}
	}
	return fmt.Errorf("%s: %w", name, ErrUnknownEventName)
}
