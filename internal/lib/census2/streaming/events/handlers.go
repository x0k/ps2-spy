package ps2events

import "github.com/x0k/ps2-spy/internal/lib/publisher"

type achievementEarnedHandler chan<- AchievementEarned

func (h achievementEarnedHandler) Type() string {
	return AchievementEarnedEventName
}

func (h achievementEarnedHandler) Handle(e publisher.Event) {
	h <- *(e.(*AchievementEarned))
}

type battleRankUpHandler chan<- BattleRankUp

func (h battleRankUpHandler) Type() string {
	return BattleRankUpEventName
}

func (h battleRankUpHandler) Handle(e publisher.Event) {
	h <- *(e.(*BattleRankUp))
}

type deathHandler chan<- Death

func (h deathHandler) Type() string {
	return DeathEventName
}

func (h deathHandler) Handle(e publisher.Event) {
	h <- *(e.(*Death))
}

type gainExperienceHandler chan<- GainExperience

func (h gainExperienceHandler) Type() string {
	return GainExperienceEventName
}

func (h gainExperienceHandler) Handle(e publisher.Event) {
	h <- *(e.(*GainExperience))
}

type itemAddedHandler chan<- ItemAdded

func (h itemAddedHandler) Type() string {
	return ItemAddedEventName
}

func (h itemAddedHandler) Handle(e publisher.Event) {
	h <- *(e.(*ItemAdded))
}

type playerFacilityCaptureHandler chan<- PlayerFacilityCapture

func (h playerFacilityCaptureHandler) Type() string {
	return PlayerFacilityCaptureEventName
}

func (h playerFacilityCaptureHandler) Handle(e publisher.Event) {
	h <- *(e.(*PlayerFacilityCapture))
}

type playerFacilityDefendHandler chan<- PlayerFacilityDefend

func (h playerFacilityDefendHandler) Type() string {
	return PlayerFacilityDefendEventName
}

func (h playerFacilityDefendHandler) Handle(e publisher.Event) {
	h <- *(e.(*PlayerFacilityDefend))
}

type playerLoginHandler chan<- PlayerLogin

func (h playerLoginHandler) Type() string {
	return PlayerLoginEventName
}

func (h playerLoginHandler) Handle(e publisher.Event) {
	h <- *(e.(*PlayerLogin))
}

type playerLogoutHandler chan<- PlayerLogout

func (h playerLogoutHandler) Type() string {
	return PlayerLogoutEventName
}

func (h playerLogoutHandler) Handle(e publisher.Event) {
	h <- *(e.(*PlayerLogout))
}

type skillAddedHandler chan<- SkillAdded

func (h skillAddedHandler) Type() string {
	return SkillAddedEventName
}

func (h skillAddedHandler) Handle(e publisher.Event) {
	h <- *(e.(*SkillAdded))
}

type vehicleDestroyHandler chan<- VehicleDestroy

func (h vehicleDestroyHandler) Type() string {
	return VehicleDestroyEventName
}

func (h vehicleDestroyHandler) Handle(e publisher.Event) {
	h <- *(e.(*VehicleDestroy))
}

type continentLockHandler chan<- ContinentLock

func (h continentLockHandler) Type() string {
	return ContinentLockEventName
}

func (h continentLockHandler) Handle(e publisher.Event) {
	h <- *(e.(*ContinentLock))
}

type facilityControlHandler chan<- FacilityControl

func (h facilityControlHandler) Type() string {
	return FacilityControlEventName
}

func (h facilityControlHandler) Handle(e publisher.Event) {
	h <- *(e.(*FacilityControl))
}

type metagameEventHandler chan<- MetagameEvent

func (h metagameEventHandler) Type() string {
	return MetagameEventEventName
}

func (h metagameEventHandler) Handle(e publisher.Event) {
	h <- *(e.(*MetagameEvent))
}
