package ps2events

import "github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"

const (
	AchievementEarnedEventName     = "AchievementEarned"
	BattleRankUpEventName          = "BattleRankUp"
	DeathEventName                 = "Death"
	GainExperienceEventName        = "GainExperience"
	ItemAddedEventName             = "ItemAdded"
	PlayerFacilityCaptureEventName = "PlayerFacilityCapture"
	PlayerFacilityDefendEventName  = "PlayerFacilityDefend"
	PlayerLoginEventName           = "PlayerLogin"
	PlayerLogoutEventName          = "PlayerLogout"
	SkillAddedEventName            = "SkillAdded"
	VehicleDestroyEventName        = "VehicleDestroy"
	ContinentLockEventName         = "ContinentLock"
	FacilityControlEventName       = "FacilityControl"
	MetagameEventEventName         = "MetagameEvent"
)

type AchievementEarned struct {
	core.EventBase
	CharacterID   string `json:"character_id"`
	WorldID       string `json:"world_id"`
	AchievementID string `json:"achievement_id"`
	ZoneID        string `json:"zone_id"`
}

type BattleRankUp struct {
	core.EventBase
	BattleRank  string `json:"battle_rank"`
	CharacterID string `json:"character_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
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

type ItemAdded struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	Context     string `json:"context"`
	ItemCount   string `json:"item_count"`
	ItemID      string `json:"item_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type PlayerFacilityCapture struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	FacilityID  string `json:"facility_id"`
	OutfitID    string `json:"outfit_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type PlayerFacilityDefend struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	FacilityID  string `json:"facility_id"`
	OutfitID    string `json:"outfit_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type PlayerLogin struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	WorldID     string `json:"world_id"`
}

type PlayerLogout struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	WorldID     string `json:"world_id"`
}

type SkillAdded struct {
	core.EventBase
	CharacterID string `json:"character_id"`
	SkillID     string `json:"skill_id"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
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
