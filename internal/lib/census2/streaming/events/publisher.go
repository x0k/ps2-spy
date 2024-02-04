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
	publisher.Publisher[publisher.Event]
	buffers map[string]any
}

func NewPublisher(publisher publisher.Publisher[publisher.Event]) *Publisher {
	return &Publisher{
		Publisher: publisher,
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
			return p.Publisher.Publish(e)
		}
	}
	return fmt.Errorf("%s: %w", name, ErrUnknownEventName)
}

func (p *Publisher) AddAchievementEarnedHandler(c chan<- AchievementEarned) func() {
	return p.AddHandler(achievementEarnedHandler(c))
}

func (p *Publisher) AddBattleRankUpHandler(c chan<- BattleRankUp) func() {
	return p.AddHandler(battleRankUpHandler(c))
}

func (p *Publisher) AddDeathHandler(c chan<- Death) func() {
	return p.AddHandler(deathHandler(c))
}

func (p *Publisher) AddGainExperienceHandler(c chan<- GainExperience) func() {
	return p.AddHandler(gainExperienceHandler(c))
}

func (p *Publisher) AddItemAddedHandler(c chan<- ItemAdded) func() {
	return p.AddHandler(itemAddedHandler(c))
}

func (p *Publisher) AddPlayerFacilityCaptureHandler(c chan<- PlayerFacilityCapture) func() {
	return p.AddHandler(playerFacilityCaptureHandler(c))
}

func (p *Publisher) AddPlayerFacilityDefendHandler(c chan<- PlayerFacilityDefend) func() {
	return p.AddHandler(playerFacilityDefendHandler(c))
}

func (p *Publisher) AddPlayerLoginHandler(c chan<- PlayerLogin) func() {
	return p.AddHandler(playerLoginHandler(c))
}

func (p *Publisher) AddPlayerLogoutHandler(c chan<- PlayerLogout) func() {
	return p.AddHandler(playerLogoutHandler(c))
}

func (p *Publisher) AddSkillAddedHandler(c chan<- SkillAdded) func() {
	return p.AddHandler(skillAddedHandler(c))
}

func (p *Publisher) AddVehicleDestroyHandler(c chan<- VehicleDestroy) func() {
	return p.AddHandler(vehicleDestroyHandler(c))
}

func (p *Publisher) AddContinentLockHandler(c chan<- ContinentLock) func() {
	return p.AddHandler(continentLockHandler(c))
}

func (p *Publisher) AddFacilityControlHandler(c chan<- FacilityControl) func() {
	return p.AddHandler(facilityControlHandler(c))
}

func (p *Publisher) AddMetagameEventHandler(c chan<- MetagameEvent) func() {
	return p.AddHandler(metagameEventHandler(c))
}
