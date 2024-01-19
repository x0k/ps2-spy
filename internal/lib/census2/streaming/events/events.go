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
	core.EventBase `mapstructure:",squash"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	AchievementID  string `json:"achievement_id" mapstructure:"achievement_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

type BattleRankUp struct {
	core.EventBase `mapstructure:",squash"`
	BattleRank     string `json:"battle_rank" mapstructure:"battle_rank"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

type Death struct {
	core.EventBase      `mapstructure:",squash"`
	AttackerCharacterID string `json:"attacker_character_id" mapstructure:"attacker_character_id"`
	AttackerFireModeID  string `json:"attacker_fire_mode_id" mapstructure:"attacker_fire_mode_id"`
	AttackerLoadoutID   string `json:"attacker_loadout_id" mapstructure:"attacker_loadout_id"`
	AttackerTeamID      string `json:"attacker_team_id" mapstructure:"attacker_team_id"`
	AttackerVehicleID   string `json:"attacker_vehicle_id" mapstructure:"attacker_vehicle_id"`
	AttackerWeaponID    string `json:"attacker_weapon_id" mapstructure:"attacker_weapon_id"`
	CharacterID         string `json:"character_id" mapstructure:"character_id"`
	CharacterLoadoutID  string `json:"character_loadout_id" mapstructure:"character_loadout_id"`
	IsCritical          string `json:"is_critical" mapstructure:"is_critical"`
	IsHeadshot          string `json:"is_headshot" mapstructure:"is_headshot"`
	TeamID              string `json:"team_id" mapstructure:"team_id"`
	WorldID             string `json:"world_id" mapstructure:"world_id"`
	ZoneID              string `json:"zone_id" mapstructure:"zone_id"`
}

type GainExperience struct {
	core.EventBase `mapstructure:",squash"`
	Amount         string `json:"amount" mapstructure:"amount"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	ExperienceID   string `json:"experience_id" mapstructure:"experience_id"`
	LoadoutID      string `json:"loadout_id" mapstructure:"loadout_id"`
	OtherID        string `json:"other_id" mapstructure:"other_id"`
	TeamID         string `json:"team_id" mapstructure:"team_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

type ItemAdded struct {
	core.EventBase `mapstructure:",squash"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	Context        string `json:"context" mapstructure:"context"`
	ItemCount      string `json:"item_count" mapstructure:"item_count"`
	ItemID         string `json:"item_id" mapstructure:"item_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

type PlayerFacilityCapture struct {
	core.EventBase `mapstructure:",squash"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	FacilityID     string `json:"facility_id" mapstructure:"facility_id"`
	OutfitID       string `json:"outfit_id" mapstructure:"outfit_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

type PlayerFacilityDefend struct {
	core.EventBase `mapstructure:",squash"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	FacilityID     string `json:"facility_id" mapstructure:"facility_id"`
	OutfitID       string `json:"outfit_id" mapstructure:"outfit_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

const CharacterIdField = "character_id"

type PlayerLogin struct {
	core.EventBase `mapstructure:",squash"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
}

type PlayerLogout struct {
	core.EventBase `mapstructure:",squash"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
}

type SkillAdded struct {
	core.EventBase `mapstructure:",squash"`
	CharacterID    string `json:"character_id" mapstructure:"character_id"`
	SkillID        string `json:"skill_id" mapstructure:"skill_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

type VehicleDestroy struct {
	core.EventBase      `mapstructure:",squash"`
	AttackerCharacterID string `json:"attacker_character_id" mapstructure:"attacker_character_id"`
	AttackerLoadoutID   string `json:"attacker_loadout_id" mapstructure:"attacker_loadout_id"`
	AttackerTeamID      string `json:"attacker_team_id" mapstructure:"attacker_team_id"`
	AttackerVehicleID   string `json:"attacker_vehicle_id" mapstructure:"attacker_vehicle_id"`
	AttackerWeaponID    string `json:"attacker_weapon_id" mapstructure:"attacker_weapon_id"`
	CharacterID         string `json:"character_id" mapstructure:"character_id"`
	FacilityID          string `json:"facility_id" mapstructure:"facility_id"`
	FactionID           string `json:"faction_id" mapstructure:"faction_id"`
	TeamID              string `json:"team_id" mapstructure:"team_id"`
	VehicleID           string `json:"vehicle_id" mapstructure:"vehicle_id"`
	WorldID             string `json:"world_id" mapstructure:"world_id"`
	ZoneID              string `json:"zone_id" mapstructure:"zone_id"`
}

type ContinentLock struct {
	core.EventBase    `mapstructure:",squash"`
	EventType         string `json:"event_type" mapstructure:"event_type"`
	MetagameEventID   string `json:"metagame_event_id" mapstructure:"metagame_event_id"`
	NCPopulation      string `json:"nc_population" mapstructure:"nc_population"`
	PreviousFaction   string `json:"previous_faction" mapstructure:"previous_faction"`
	TRPopulation      string `json:"tr_population" mapstructure:"tr_population"`
	TriggeringFaction string `json:"triggering_faction" mapstructure:"triggering_faction"`
	VSPopulation      string `json:"vs_population" mapstructure:"vs_population"`
	WorldID           string `json:"world_id" mapstructure:"world_id"`
	ZoneID            string `json:"zone_id" mapstructure:"zone_id"`
}

type FacilityControl struct {
	core.EventBase `mapstructure:",squash"`
	DurationHeld   string `json:"duration_held" mapstructure:"duration_held"`
	FacilityID     string `json:"facility_id" mapstructure:"facility_id"`
	NewFactionID   string `json:"new_faction_id" mapstructure:"new_faction_id"`
	OldFactionID   string `json:"old_faction_id" mapstructure:"old_faction_id"`
	OutfitID       string `json:"outfit_id" mapstructure:"outfit_id"`
	WorldID        string `json:"world_id" mapstructure:"world_id"`
	ZoneID         string `json:"zone_id" mapstructure:"zone_id"`
}

type MetagameEvent struct {
	core.EventBase         `mapstructure:",squash"`
	ExperienceBonus        string `json:"experience_bonus" mapstructure:"experience_bonus"`
	FactionNC              string `json:"faction_nc" mapstructure:"faction_nc"`
	FactionTR              string `json:"faction_tr" mapstructure:"faction_tr"`
	FactionVS              string `json:"faction_vs" mapstructure:"faction_vs"`
	MetagameEventID        string `json:"metagame_event_id" mapstructure:"metagame_event_id"`
	MetagameEventState     string `json:"metagame_event_state" mapstructure:"metagame_event_state"`
	MetagameEventStateName string `json:"metagame_event_state_name" mapstructure:"metagame_event_state_name"`
	WorldID                string `json:"world_id" mapstructure:"world_id"`
	InstanceID             string `json:"instance_id" mapstructure:"instance_id"`
	ZoneID                 string `json:"zone_id" mapstructure:"zone_id"`
}
