package events

import (
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	AchievementEarnedEventName     EventType = "AchievementEarned"
	BattleRankUpEventName          EventType = "BattleRankUp"
	DeathEventName                 EventType = "Death"
	GainExperienceEventName        EventType = "GainExperience"
	ItemAddedEventName             EventType = "ItemAdded"
	PlayerFacilityCaptureEventName EventType = "PlayerFacilityCapture"
	PlayerFacilityDefendEventName  EventType = "PlayerFacilityDefend"
	PlayerLoginEventName           EventType = "PlayerLogin"
	PlayerLogoutEventName          EventType = "PlayerLogout"
	SkillAddedEventName            EventType = "SkillAdded"
	VehicleDestroyEventName        EventType = "VehicleDestroy"
	ContinentLockEventName         EventType = "ContinentLock"
	FacilityControlEventName       EventType = "FacilityControl"
	MetagameEventEventName         EventType = "MetagameEvent"
)

type AchievementEarned struct {
	core.EventBase
	CharacterID   string `json:"character_id"`
	WorldID       string `json:"world_id"`
	AchievementID string `json:"achievement_id"`
	ZoneID        string `json:"zone_id"`
}

func (AchievementEarned) Type() EventType {
	return AchievementEarnedEventName
}

type BattleRankUp struct {
	core.EventBase
	BattleRank  string `json:"battle_rank"`
	CharacterID string `json:"character_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

func (BattleRankUp) Type() EventType {
	return BattleRankUpEventName
}

type Death struct {
	core.EventBase
	AttackerCharacterID string `json:"attacker_character_id"`
	AttackerFireModeID  string `json:"attacker_fire_mode_id"`
	AttackerLoadoutID   string `json:"attacker_loadout_id"`
	AttackerTeamID      string `json:"attacker_team_id"`
	AttackerVehicleID   string `json:"attacker_vehicle_id"`
	AttackerWeaponID    string `json:"attacker_weapon_id"`
	CharacterID         string `json:"character_id"`
	CharacterLoadoutID  string `json:"character_loadout_id"`
	IsCritical          string `json:"is_critical"`
	IsHeadshot          string `json:"is_headshot"`
	TeamID              string `json:"team_id"`
	WorldID             string `json:"world_id"`
	ZoneID              string `json:"zone_id"`
}

func (Death) Type() EventType {
	return DeathEventName
}

type GainExperience struct {
	core.EventBase
	Amount       string `json:"amount"`
	CharacterID  string `json:"character_id"`
	ExperienceID string `json:"experience_id"`
	LoadoutID    string `json:"loadout_id"`
	OtherID      string `json:"other_id"`
	TeamID       string `json:"team_id"`
	WorldID      string `json:"world_id"`
	ZoneID       string `json:"zone_id"`
}

func (GainExperience) Type() EventType {
	return GainExperienceEventName
}

type ItemAdded struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	Context     string `json:"context"`
	ItemCount   string `json:"item_count"`
	ItemID      string `json:"item_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

func (ItemAdded) Type() EventType {
	return ItemAddedEventName
}

type PlayerFacilityCapture struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	FacilityID  string `json:"facility_id"`
	OutfitID    string `json:"outfit_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

func (PlayerFacilityCapture) Type() EventType {
	return PlayerFacilityCaptureEventName
}

type PlayerFacilityDefend struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	FacilityID  string `json:"facility_id"`
	OutfitID    string `json:"outfit_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

func (PlayerFacilityDefend) Type() EventType {
	return PlayerFacilityDefendEventName
}

const CharacterIdField = "character_id"

type PlayerLogin struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	WorldID     string `json:"world_id"`
}

func (PlayerLogin) Type() EventType {
	return PlayerLoginEventName
}

type PlayerLogout struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	WorldID     string `json:"world_id"`
}

func (PlayerLogout) Type() EventType {
	return PlayerLogoutEventName
}

type SkillAdded struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	SkillID     string `json:"skill_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

func (SkillAdded) Type() EventType {
	return SkillAddedEventName
}

type VehicleDestroy struct {
	core.EventBase
	AttackerCharacterID string `json:"attacker_character_id"`
	AttackerLoadoutID   string `json:"attacker_loadout_id"`
	AttackerTeamID      string `json:"attacker_team_id"`
	AttackerVehicleID   string `json:"attacker_vehicle_id"`
	AttackerWeaponID    string `json:"attacker_weapon_id"`
	CharacterID         string `json:"character_id"`
	FacilityID          string `json:"facility_id"`
	FactionID           string `json:"faction_id"`
	TeamID              string `json:"team_id"`
	VehicleID           string `json:"vehicle_id"`
	WorldID             string `json:"world_id"`
	ZoneID              string `json:"zone_id"`
}

func (VehicleDestroy) Type() EventType {
	return VehicleDestroyEventName
}

type ContinentLock struct {
	core.EventBase
	EventType         string `json:"event_type"`
	MetagameEventID   string `json:"metagame_event_id"`
	NCPopulation      string `json:"nc_population"`
	PreviousFaction   string `json:"previous_faction"`
	TRPopulation      string `json:"tr_population"`
	TriggeringFaction string `json:"triggering_faction"`
	VSPopulation      string `json:"vs_population"`
	WorldID           string `json:"world_id"`
	ZoneID            string `json:"zone_id"`
}

func (ContinentLock) Type() EventType {
	return ContinentLockEventName
}

type FacilityControl struct {
	core.EventBase
	DurationHeld string `json:"duration_held"`
	FacilityID   string `json:"facility_id"`
	NewFactionID string `json:"new_faction_id"`
	OldFactionID string `json:"old_faction_id"`
	OutfitID     string `json:"outfit_id"`
	WorldID      string `json:"world_id"`
	ZoneID       string `json:"zone_id"`
}

func (FacilityControl) Type() EventType {
	return FacilityControlEventName
}

type MetagameEvent struct {
	core.EventBase
	ExperienceBonus        string `json:"experience_bonus"`
	FactionNC              string `json:"faction_nc"`
	FactionTR              string `json:"faction_tr"`
	FactionVS              string `json:"faction_vs"`
	MetagameEventID        string `json:"metagame_event_id"`
	MetagameEventState     string `json:"metagame_event_state"`
	MetagameEventStateName string `json:"metagame_event_state_name"`
	WorldID                string `json:"world_id"`
	InstanceID             string `json:"instance_id"`
	ZoneID                 string `json:"zone_id"`
}

func (MetagameEvent) Type() EventType {
	return MetagameEventEventName
}
