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
	if t, ok := e.(*AchievementEarned); ok {
		h <- *t
	}
}

type battleRankUpHandler chan<- BattleRankUp

func (h battleRankUpHandler) Type() string {
	return BattleRankUpEventName
}

func (h battleRankUpHandler) Handle(e any) {
	if t, ok := e.(*BattleRankUp); ok {
		h <- *t
	}
}

type deathHandler chan<- Death

func (h deathHandler) Type() string {
	return DeathEventName
}

func (h deathHandler) Handle(e any) {
	if t, ok := e.(*Death); ok {
		h <- *t
	}
}

type gainExperienceHandler chan<- GainExperience

func (h gainExperienceHandler) Type() string {
	return GainExperienceEventName
}

func (h gainExperienceHandler) Handle(e any) {
	if t, ok := e.(*GainExperience); ok {
		h <- *t
	}
}

type itemAddedHandler chan<- ItemAdded

func (h itemAddedHandler) Type() string {
	return ItemAddedEventName
}

func (h itemAddedHandler) Handle(e any) {
	if t, ok := e.(*ItemAdded); ok {
		h <- *t
	}
}

type playerFacilityCaptureHandler chan<- PlayerFacilityCapture

func (h playerFacilityCaptureHandler) Type() string {
	return PlayerFacilityCaptureEventName
}

func (h playerFacilityCaptureHandler) Handle(e any) {
	if t, ok := e.(*PlayerFacilityCapture); ok {
		h <- *t
	}
}

type playerFacilityDefendHandler chan<- PlayerFacilityDefend

func (h playerFacilityDefendHandler) Type() string {
	return PlayerFacilityDefendEventName
}

func (h playerFacilityDefendHandler) Handle(e any) {
	if t, ok := e.(*PlayerFacilityDefend); ok {
		h <- *t
	}
}

type playerLoginHandler chan<- PlayerLogin

func (h playerLoginHandler) Type() string {
	return PlayerLoginEventName
}

func (h playerLoginHandler) Handle(e any) {
	if t, ok := e.(*PlayerLogin); ok {
		h <- *t
	}
}

type playerLogoutHandler chan<- PlayerLogout

func (h playerLogoutHandler) Type() string {
	return PlayerLogoutEventName
}

func (h playerLogoutHandler) Handle(e any) {
	if t, ok := e.(*PlayerLogout); ok {
		h <- *t
	}
}

type skillAddedHandler chan<- SkillAdded

func (h skillAddedHandler) Type() string {
	return SkillAddedEventName
}

func (h skillAddedHandler) Handle(e any) {
	if t, ok := e.(*SkillAdded); ok {
		h <- *t
	}
}

type vehicleDestroyHandler chan<- VehicleDestroy

func (h vehicleDestroyHandler) Type() string {
	return VehicleDestroyEventName
}

func (h vehicleDestroyHandler) Handle(e any) {
	if t, ok := e.(*VehicleDestroy); ok {
		h <- *t
	}
}

type continentLockHandler chan<- ContinentLock

func (h continentLockHandler) Type() string {
	return ContinentLockEventName
}

func (h continentLockHandler) Handle(e any) {
	if t, ok := e.(*ContinentLock); ok {
		h <- *t
	}
}

type facilityControlHandler chan<- FacilityControl

func (h facilityControlHandler) Type() string {
	return FacilityControlEventName
}

func (h facilityControlHandler) Handle(e any) {
	if t, ok := e.(*FacilityControl); ok {
		h <- *t
	}
}

type metagameEventHandler chan<- MetagameEvent

func (h metagameEventHandler) Type() string {
	return MetagameEventEventName
}

func (h metagameEventHandler) Handle(e any) {
	if t, ok := e.(*MetagameEvent); ok {
		h <- *t
	}
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
