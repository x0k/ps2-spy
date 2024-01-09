package ps2events

type EventHandler interface {
	Type() string
	Handle(e any)
}

type achievementEarnedHandler func(e *AchievementEarned)

func (h achievementEarnedHandler) Type() string {
	return AchievementEarnedEventName
}

func (h achievementEarnedHandler) Handle(e any) {
	if t, ok := e.(*AchievementEarned); ok {
		h(t)
	}
}

type battleRankUpHandler func(e *BattleRankUp)

func (h battleRankUpHandler) Type() string {
	return BattleRankUpEventName
}

func (h battleRankUpHandler) Handle(e any) {
	if t, ok := e.(*BattleRankUp); ok {
		h(t)
	}
}

type deathHandler func(e *Death)

func (h deathHandler) Type() string {
	return DeathEventName
}

func (h deathHandler) Handle(e any) {
	if t, ok := e.(*Death); ok {
		h(t)
	}
}

type gainExperienceHandler func(e *GainExperience)

func (h gainExperienceHandler) Type() string {
	return GainExperienceEventName
}

func (h gainExperienceHandler) Handle(e any) {
	if t, ok := e.(*GainExperience); ok {
		h(t)
	}
}

type itemAddedHandler func(e *ItemAdded)

func (h itemAddedHandler) Type() string {
	return ItemAddedEventName
}

func (h itemAddedHandler) Handle(e any) {
	if t, ok := e.(*ItemAdded); ok {
		h(t)
	}
}

type playerFacilityCaptureHandler func(e *PlayerFacilityCapture)

func (h playerFacilityCaptureHandler) Type() string {
	return PlayerFacilityCaptureEventName
}

func (h playerFacilityCaptureHandler) Handle(e any) {
	if t, ok := e.(*PlayerFacilityCapture); ok {
		h(t)
	}
}

type playerFacilityDefendHandler func(e *PlayerFacilityDefend)

func (h playerFacilityDefendHandler) Type() string {
	return PlayerFacilityDefendEventName
}

func (h playerFacilityDefendHandler) Handle(e any) {
	if t, ok := e.(*PlayerFacilityDefend); ok {
		h(t)
	}
}

type playerLoginHandler func(e *PlayerLogin)

func (h playerLoginHandler) Type() string {
	return PlayerLoginEventName
}

func (h playerLoginHandler) Handle(e any) {
	if t, ok := e.(*PlayerLogin); ok {
		h(t)
	}
}

type playerLogoutHandler func(e *PlayerLogout)

func (h playerLogoutHandler) Type() string {
	return PlayerLogoutEventName
}

func (h playerLogoutHandler) Handle(e any) {
	if t, ok := e.(*PlayerLogout); ok {
		h(t)
	}
}

type skillAddedHandler func(e *SkillAdded)

func (h skillAddedHandler) Type() string {
	return SkillAddedEventName
}

func (h skillAddedHandler) Handle(e any) {
	if t, ok := e.(*SkillAdded); ok {
		h(t)
	}
}

type vehicleDestroyHandler func(e *VehicleDestroy)

func (h vehicleDestroyHandler) Type() string {
	return VehicleDestroyEventName
}

func (h vehicleDestroyHandler) Handle(e any) {
	if t, ok := e.(*VehicleDestroy); ok {
		h(t)
	}
}

type continentLockHandler func(e *ContinentLock)

func (h continentLockHandler) Type() string {
	return ContinentLockEventName
}

func (h continentLockHandler) Handle(e any) {
	if t, ok := e.(*ContinentLock); ok {
		h(t)
	}
}

type facilityControlHandler func(e *FacilityControl)

func (h facilityControlHandler) Type() string {
	return FacilityControlEventName
}

func (h facilityControlHandler) Handle(e any) {
	if t, ok := e.(*FacilityControl); ok {
		h(t)
	}
}

type metagameEventHandler func(e *MetagameEvent)

func (h metagameEventHandler) Type() string {
	return MetagameEventEventName
}

func (h metagameEventHandler) Handle(e any) {
	if t, ok := e.(*MetagameEvent); ok {
		h(t)
	}
}

type anyEventHandler func(e any)

func (h anyEventHandler) Type() string {
	return "all"
}

func (h anyEventHandler) Handle(e any) {
	h(e)
}

func EventHandlerForInterface(handler interface{}) EventHandler {
	switch v := handler.(type) {
	case func(e any):
		return anyEventHandler(v)
	case func(e *AchievementEarned):
		return achievementEarnedHandler(v)
	case func(e *BattleRankUp):
		return battleRankUpHandler(v)
	case func(e *Death):
		return deathHandler(v)
	case func(e *GainExperience):
		return gainExperienceHandler(v)
	case func(e *ItemAdded):
		return itemAddedHandler(v)
	case func(e *PlayerFacilityCapture):
		return playerFacilityCaptureHandler(v)
	case func(e *PlayerFacilityDefend):
		return playerFacilityDefendHandler(v)
	case func(e *PlayerLogin):
		return playerLoginHandler(v)
	case func(e *PlayerLogout):
		return playerLogoutHandler(v)
	case func(e *SkillAdded):
		return skillAddedHandler(v)
	case func(e *VehicleDestroy):
		return vehicleDestroyHandler(v)
	case func(e *ContinentLock):
		return continentLockHandler(v)
	case func(e *FacilityControl):
		return facilityControlHandler(v)
	case func(e *MetagameEvent):
		return metagameEventHandler(v)
	}
	return nil
}
