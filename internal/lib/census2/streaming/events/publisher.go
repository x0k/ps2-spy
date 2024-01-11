package ps2events

import (
	"fmt"
	"log/slog"

	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

var ErrEventNameNotFound = fmt.Errorf("event name not found")
var ErrUnknownEventName = fmt.Errorf("unknown event name")

type Publisher struct {
	log                       *slog.Logger
	achievementEarnedBuff     AchievementEarned
	AchievementEarned         chan AchievementEarned
	battleRankUpBuff          BattleRankUp
	BattleRankUp              chan BattleRankUp
	deathBuff                 Death
	Death                     chan Death
	gainExperienceBuff        GainExperience
	GainExperience            chan GainExperience
	itemAddedBuff             ItemAdded
	ItemAdded                 chan ItemAdded
	playerFacilityCaptureBuff PlayerFacilityCapture
	PlayerFacilityCapture     chan PlayerFacilityCapture
	playerFacilityDefendBuff  PlayerFacilityDefend
	PlayerFacilityDefend      chan PlayerFacilityDefend
	playerLoginBuff           PlayerLogin
	PlayerLogin               chan PlayerLogin
	playerLogoutBuff          PlayerLogout
	PlayerLogout              chan PlayerLogout
	skillAddedBuff            SkillAdded
	SkillAdded                chan SkillAdded
	vehicleDestroyBuff        VehicleDestroy
	VehicleDestroy            chan VehicleDestroy
	continentLockBuff         ContinentLock
	ContinentLock             chan ContinentLock
	facilityControlBuff       FacilityControl
	FacilityControl           chan FacilityControl
	metagameEventBuff         MetagameEvent
	MetagameEvent             chan MetagameEvent
}

func NewPublisher(log *slog.Logger) *Publisher {
	return &Publisher{
		log: log.With(
			slog.String("component", "census2.streaming.events.Publisher"),
		),
		AchievementEarned: make(chan AchievementEarned),
		BattleRankUp:      make(chan BattleRankUp),
		Death:             make(chan Death),
		GainExperience:    make(chan GainExperience),
		ItemAdded:         make(chan ItemAdded),
		PlayerLogin:       make(chan PlayerLogin),
		PlayerLogout:      make(chan PlayerLogout),
		SkillAdded:        make(chan SkillAdded),
		VehicleDestroy:    make(chan VehicleDestroy),
		ContinentLock:     make(chan ContinentLock),
		FacilityControl:   make(chan FacilityControl),
		MetagameEvent:     make(chan MetagameEvent),
	}
}

func (p *Publisher) Publish(event map[string]any) {
	var err error
	defer func() {
		if err != nil {
			p.log.Warn("failed to publish event", slog.Any("event", event), sl.Err(err))
		}
	}()
	name, ok := event[core.EventNameField]
	if !ok {
		err = ErrEventNameNotFound
		return
	}
	switch name {
	case AchievementEarnedEventName:
		err = mapstructure.Decode(event, &p.achievementEarnedBuff)
		if err != nil {
			return
		}
		p.AchievementEarned <- p.achievementEarnedBuff
	case BattleRankUpEventName:
		err = mapstructure.Decode(event, &p.battleRankUpBuff)
		if err != nil {
			return
		}
		p.BattleRankUp <- p.battleRankUpBuff
	case DeathEventName:
		err = mapstructure.Decode(event, &p.deathBuff)
		if err != nil {
			return
		}
		p.Death <- p.deathBuff
	case GainExperienceEventName:
		err = mapstructure.Decode(event, &p.gainExperienceBuff)
		if err != nil {
			return
		}
		p.GainExperience <- p.gainExperienceBuff
	case ItemAddedEventName:
		err = mapstructure.Decode(event, &p.itemAddedBuff)
		if err != nil {
			return
		}
		p.ItemAdded <- p.itemAddedBuff
	case PlayerFacilityCaptureEventName:
		err = mapstructure.Decode(event, &p.playerFacilityCaptureBuff)
		if err != nil {
			return
		}
		p.PlayerFacilityCapture <- p.playerFacilityCaptureBuff
	case PlayerFacilityDefendEventName:
		err = mapstructure.Decode(event, &p.playerFacilityDefendBuff)
		if err != nil {
			return
		}
		p.PlayerFacilityDefend <- p.playerFacilityDefendBuff
	case PlayerLoginEventName:
		err = mapstructure.Decode(event, &p.playerLoginBuff)
		if err != nil {
			return
		}
		p.PlayerLogin <- p.playerLoginBuff
	case PlayerLogoutEventName:
		err = mapstructure.Decode(event, &p.playerLogoutBuff)
		if err != nil {
			return
		}
		p.PlayerLogout <- p.playerLogoutBuff
	case SkillAddedEventName:
		err = mapstructure.Decode(event, &p.skillAddedBuff)
		if err != nil {
			return
		}
		p.SkillAdded <- p.skillAddedBuff
	case VehicleDestroyEventName:
		err = mapstructure.Decode(event, &p.vehicleDestroyBuff)
		if err != nil {
			return
		}
		p.VehicleDestroy <- p.vehicleDestroyBuff
	case ContinentLockEventName:
		err = mapstructure.Decode(event, &p.continentLockBuff)
		if err != nil {
			return
		}
		p.ContinentLock <- p.continentLockBuff
	case FacilityControlEventName:
		err = mapstructure.Decode(event, &p.facilityControlBuff)
		if err != nil {
			return
		}
		p.FacilityControl <- p.facilityControlBuff
	case MetagameEventEventName:
		err = mapstructure.Decode(event, &p.metagameEventBuff)
		if err != nil {
			return
		}
		p.MetagameEvent <- p.metagameEventBuff
	default:
		err = fmt.Errorf("%s: %w", name, ErrUnknownEventName)
	}
}
