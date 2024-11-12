package events

type achievementEarnedHandler chan<- AchievementEarned

func (h achievementEarnedHandler) Type() EventType {
	return AchievementEarnedEventName
}

func (h achievementEarnedHandler) Handle(e Event) {
	h <- *(e.(*AchievementEarned))
}

type battleRankUpHandler chan<- BattleRankUp

func (h battleRankUpHandler) Type() EventType {
	return BattleRankUpEventName
}

func (h battleRankUpHandler) Handle(e Event) {
	h <- *(e.(*BattleRankUp))
}

type deathHandler chan<- Death

func (h deathHandler) Type() EventType {
	return DeathEventName
}

func (h deathHandler) Handle(e Event) {
	h <- *(e.(*Death))
}

type gainExperienceHandler chan<- GainExperience

func (h gainExperienceHandler) Type() EventType {
	return GainExperienceEventName
}

func (h gainExperienceHandler) Handle(e Event) {
	h <- *(e.(*GainExperience))
}

type itemAddedHandler chan<- ItemAdded

func (h itemAddedHandler) Type() EventType {
	return ItemAddedEventName
}

func (h itemAddedHandler) Handle(e Event) {
	h <- *(e.(*ItemAdded))
}

type playerFacilityCaptureHandler chan<- PlayerFacilityCapture

func (h playerFacilityCaptureHandler) Type() EventType {
	return PlayerFacilityCaptureEventName
}

func (h playerFacilityCaptureHandler) Handle(e Event) {
	h <- *(e.(*PlayerFacilityCapture))
}

type playerFacilityDefendHandler chan<- PlayerFacilityDefend

func (h playerFacilityDefendHandler) Type() EventType {
	return PlayerFacilityDefendEventName
}

func (h playerFacilityDefendHandler) Handle(e Event) {
	h <- *(e.(*PlayerFacilityDefend))
}

type playerLoginHandler chan<- PlayerLogin

func (h playerLoginHandler) Type() EventType {
	return PlayerLoginEventName
}

func (h playerLoginHandler) Handle(e Event) {
	h <- *(e.(*PlayerLogin))
}

type playerLogoutHandler chan<- PlayerLogout

func (h playerLogoutHandler) Type() EventType {
	return PlayerLogoutEventName
}

func (h playerLogoutHandler) Handle(e Event) {
	h <- *(e.(*PlayerLogout))
}

type skillAddedHandler chan<- SkillAdded

func (h skillAddedHandler) Type() EventType {
	return SkillAddedEventName
}

func (h skillAddedHandler) Handle(e Event) {
	h <- *(e.(*SkillAdded))
}

type vehicleDestroyHandler chan<- VehicleDestroy

func (h vehicleDestroyHandler) Type() EventType {
	return VehicleDestroyEventName
}

func (h vehicleDestroyHandler) Handle(e Event) {
	h <- *(e.(*VehicleDestroy))
}

type continentLockHandler chan<- ContinentLock

func (h continentLockHandler) Type() EventType {
	return ContinentLockEventName
}

func (h continentLockHandler) Handle(e Event) {
	h <- *(e.(*ContinentLock))
}

type facilityControlHandler chan<- FacilityControl

func (h facilityControlHandler) Type() EventType {
	return FacilityControlEventName
}

func (h facilityControlHandler) Handle(e Event) {
	h <- *(e.(*FacilityControl))
}

type metagameEventHandler chan<- MetagameEvent

func (h metagameEventHandler) Type() EventType {
	return MetagameEventEventName
}

func (h metagameEventHandler) Handle(e Event) {
	h <- *(e.(*MetagameEvent))
}
