package ps2events

type eventHandler interface {
	Type() string
	Handle(e any)
}

type achievementEarnedHandler chan<- AchievementEarned

func (h achievementEarnedHandler) Type() string {
	return AchievementEarnedEventName
}

func (h achievementEarnedHandler) Handle(e any) {
	h <- *(e.(*AchievementEarned))
}

type battleRankUpHandler chan<- BattleRankUp

func (h battleRankUpHandler) Type() string {
	return BattleRankUpEventName
}

func (h battleRankUpHandler) Handle(e any) {
	h <- *(e.(*BattleRankUp))
}

type deathHandler chan<- Death

func (h deathHandler) Type() string {
	return DeathEventName
}

func (h deathHandler) Handle(e any) {
	h <- *(e.(*Death))
}

type gainExperienceHandler chan<- GainExperience

func (h gainExperienceHandler) Type() string {
	return GainExperienceEventName
}

func (h gainExperienceHandler) Handle(e any) {
	h <- *(e.(*GainExperience))
}

type itemAddedHandler chan<- ItemAdded

func (h itemAddedHandler) Type() string {
	return ItemAddedEventName
}

func (h itemAddedHandler) Handle(e any) {
	h <- *(e.(*ItemAdded))
}

type playerFacilityCaptureHandler chan<- PlayerFacilityCapture

func (h playerFacilityCaptureHandler) Type() string {
	return PlayerFacilityCaptureEventName
}

func (h playerFacilityCaptureHandler) Handle(e any) {
	h <- *(e.(*PlayerFacilityCapture))
}

type playerFacilityDefendHandler chan<- PlayerFacilityDefend

func (h playerFacilityDefendHandler) Type() string {
	return PlayerFacilityDefendEventName
}

func (h playerFacilityDefendHandler) Handle(e any) {
	h <- *(e.(*PlayerFacilityDefend))
}

type playerLoginHandler chan<- PlayerLogin

func (h playerLoginHandler) Type() string {
	return PlayerLoginEventName
}

func (h playerLoginHandler) Handle(e any) {
	h <- *(e.(*PlayerLogin))
}

type playerLogoutHandler chan<- PlayerLogout

func (h playerLogoutHandler) Type() string {
	return PlayerLogoutEventName
}

func (h playerLogoutHandler) Handle(e any) {
	h <- *(e.(*PlayerLogout))
}

type skillAddedHandler chan<- SkillAdded

func (h skillAddedHandler) Type() string {
	return SkillAddedEventName
}

func (h skillAddedHandler) Handle(e any) {
	h <- *(e.(*SkillAdded))
}

type vehicleDestroyHandler chan<- VehicleDestroy

func (h vehicleDestroyHandler) Type() string {
	return VehicleDestroyEventName
}

func (h vehicleDestroyHandler) Handle(e any) {
	h <- *(e.(*VehicleDestroy))
}

type continentLockHandler chan<- ContinentLock

func (h continentLockHandler) Type() string {
	return ContinentLockEventName
}

func (h continentLockHandler) Handle(e any) {
	h <- *(e.(*ContinentLock))
}

type facilityControlHandler chan<- FacilityControl

func (h facilityControlHandler) Type() string {
	return FacilityControlEventName
}

func (h facilityControlHandler) Handle(e any) {
	h <- *(e.(*FacilityControl))
}

type metagameEventHandler chan<- MetagameEvent

func (h metagameEventHandler) Type() string {
	return MetagameEventEventName
}

func (h metagameEventHandler) Handle(e any) {
	h <- *(e.(*MetagameEvent))
}

func handlerForInterface(handler any) eventHandler {
	switch v := handler.(type) {
	case chan AchievementEarned:
		return achievementEarnedHandler(v)
	case chan<- AchievementEarned:
		return achievementEarnedHandler(v)
	case chan BattleRankUp:
		return battleRankUpHandler(v)
	case chan<- BattleRankUp:
		return battleRankUpHandler(v)
	case chan Death:
		return deathHandler(v)
	case chan<- Death:
		return deathHandler(v)
	case chan GainExperience:
		return gainExperienceHandler(v)
	case chan<- GainExperience:
		return gainExperienceHandler(v)
	case chan ItemAdded:
		return itemAddedHandler(v)
	case chan<- ItemAdded:
		return itemAddedHandler(v)
	case chan PlayerFacilityCapture:
		return playerFacilityCaptureHandler(v)
	case chan<- PlayerFacilityCapture:
		return playerFacilityCaptureHandler(v)
	case chan PlayerFacilityDefend:
		return playerFacilityDefendHandler(v)
	case chan<- PlayerFacilityDefend:
		return playerFacilityDefendHandler(v)
	case chan PlayerLogin:
		return playerLoginHandler(v)
	case chan<- PlayerLogin:
		return playerLoginHandler(v)
	case chan PlayerLogout:
		return playerLogoutHandler(v)
	case chan<- PlayerLogout:
		return playerLogoutHandler(v)
	case chan SkillAdded:
		return skillAddedHandler(v)
	case chan<- SkillAdded:
		return skillAddedHandler(v)
	case chan VehicleDestroy:
		return vehicleDestroyHandler(v)
	case chan<- VehicleDestroy:
		return vehicleDestroyHandler(v)
	case chan ContinentLock:
		return continentLockHandler(v)
	case chan<- ContinentLock:
		return continentLockHandler(v)
	case chan FacilityControl:
		return facilityControlHandler(v)
	case chan<- FacilityControl:
		return facilityControlHandler(v)
	case chan MetagameEvent:
		return metagameEventHandler(v)
	case chan<- MetagameEvent:
		return metagameEventHandler(v)
	}
	return nil
}
