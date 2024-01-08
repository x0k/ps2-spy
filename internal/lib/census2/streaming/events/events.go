package ps2events

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
	EventName     string `json:"event_name"`
	CharacterID   string `json:"character_id"`
	Timestamp     string `json:"timestamp"`
	WorldID       string `json:"world_id"`
	AchievementID string `json:"achievement_id"`
	ZoneID        string `json:"zone_id"`
}

type BattleRankUp struct {
	BattleRank  string `json:"battle_rank"`
	CharacterID string `json:"character_id"`
	EventName   string `json:"event_name"`
	Timestamp   string `json:"timestamp"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type Death struct {
	AttackerCharacterID string `json:"attacker_character_id"`
	AttackerFireModeID  string `json:"attacker_fire_mode_id"`
	AttackerLoadoutID   string `json:"attacker_loadout_id"`
	AttackerTeamID      string `json:"attacker_team_id"`
	AttackerVehicleID   string `json:"attacker_vehicle_id"`
	AttackerWeaponID    string `json:"attacker_weapon_id"`
	CharacterID         string `json:"character_id"`
	CharacterLoadoutID  string `json:"character_loadout_id"`
	EventName           string `json:"event_name"`
	IsCritical          string `json:"is_critical"`
	IsHeadshot          string `json:"is_headshot"`
	TeamID              string `json:"team_id"`
	Timestamp           string `json:"timestamp"`
	WorldID             string `json:"world_id"`
	ZoneID              string `json:"zone_id"`
}

type GainExperience struct {
	Amount       string `json:"amount"`
	CharacterID  string `json:"character_id"`
	EventName    string `json:"event_name"`
	ExperienceID string `json:"experience_id"`
	LoadoutID    string `json:"loadout_id"`
	OtherID      string `json:"other_id"`
	TeamID       string `json:"team_id"`
	Timestamp    string `json:"timestamp"`
	WorldID      string `json:"world_id"`
	ZoneID       string `json:"zone_id"`
}

type ItemAdded struct {
	CharacterID string `json:"character_id"`
	Context     string `json:"context"`
	EventName   string `json:"event_name"`
	ItemCount   string `json:"item_count"`
	ItemID      string `json:"item_id"`
	Timestamp   string `json:"timestamp"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type PlayerFacilityCapture struct {
	CharacterID string `json:"character_id"`
	EventName   string `json:"event_name"`
	FacilityID  string `json:"facility_id"`
	OutfitID    string `json:"outfit_id"`
	Timestamp   string `json:"timestamp"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type PlayerFacilityDefend struct {
	CharacterID string `json:"character_id"`
	EventName   string `json:"event_name"`
	FacilityID  string `json:"facility_id"`
	OutfitID    string `json:"outfit_id"`
	Timestamp   string `json:"timestamp"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type PlayerLogin struct {
	CharacterID string `json:"character_id"`
	EventName   string `json:"event_name"`
	Timestamp   string `json:"timestamp"`
	WorldID     string `json:"world_id"`
}

type PlayerLogout struct {
	CharacterID string `json:"character_id"`
	EventName   string `json:"event_name"`
	Timestamp   string `json:"timestamp"`
	WorldID     string `json:"world_id"`
}

type SkillAdded struct {
	CharacterID string `json:"character_id"`
	EventName   string `json:"event_name"`
	SkillID     string `json:"skill_id"`
	Timestamp   string `json:"timestamp"`
	WorldID     string `json:"world_id"`
	ZoneID      string `json:"zone_id"`
}

type VehicleDestroy struct {
	AttackerCharacterID string `json:"attacker_character_id"`
	AttackerLoadoutID   string `json:"attacker_loadout_id"`
	AttackerTeamID      string `json:"attacker_team_id"`
	AttackerVehicleID   string `json:"attacker_vehicle_id"`
	AttackerWeaponID    string `json:"attacker_weapon_id"`
	CharacterID         string `json:"character_id"`
	EventName           string `json:"event_name"`
	FacilityID          string `json:"facility_id"`
	FactionID           string `json:"faction_id"`
	TeamID              string `json:"team_id"`
	Timestamp           string `json:"timestamp"`
	VehicleID           string `json:"vehicle_id"`
	WorldID             string `json:"world_id"`
	ZoneID              string `json:"zone_id"`
}

type ContinentLock struct {
	EventName         string `json:"event_name"`
	EventType         string `json:"event_type"`
	MetagameEventID   string `json:"metagame_event_id"`
	NCPopulation      string `json:"nc_population"`
	PreviousFaction   string `json:"previous_faction"`
	Timestamp         string `json:"timestamp"`
	TRPopulation      string `json:"tr_population"`
	TriggeringFaction string `json:"triggering_faction"`
	VSPopulation      string `json:"vs_population"`
	WorldID           string `json:"world_id"`
	ZoneID            string `json:"zone_id"`
}

type FacilityControl struct {
	DurationHeld string `json:"duration_held"`
	EventName    string `json:"event_name"`
	FacilityID   string `json:"facility_id"`
	NewFactionID string `json:"new_faction_id"`
	OldFactionID string `json:"old_faction_id"`
	OutfitID     string `json:"outfit_id"`
	Timestamp    string `json:"timestamp"`
	WorldID      string `json:"world_id"`
	ZoneID       string `json:"zone_id"`
}

type MetagameEvent struct {
	EventName              string `json:"event_name"`
	ExperienceBonus        string `json:"experience_bonus"`
	FactionNC              string `json:"faction_nc"`
	FactionTR              string `json:"faction_tr"`
	FactionVS              string `json:"faction_vs"`
	MetagameEventID        string `json:"metagame_event_id"`
	MetagameEventState     string `json:"metagame_event_state"`
	MetagameEventStateName string `json:"metagame_event_state_name"`
	Timestamp              string `json:"timestamp"`
	WorldID                string `json:"world_id"`
	InstanceID             string `json:"instance_id"`
	ZoneID                 string `json:"zone_id"`
}
