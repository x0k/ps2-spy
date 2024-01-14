package collections

const WorldEvent = "world_event"

type WorldEventItem struct {
	MetagameEventId        string `json:"metagame_event_id" mapstructure:"metagame_event_id"`
	MetagameEventState     string `json:"metagame_event_state" mapstructure:"metagame_event_state"`
	FactionNC              string `json:"faction_nc" mapstructure:"faction_nc"`
	FactionTR              string `json:"faction_tr" mapstructure:"faction_tr"`
	FactionVS              string `json:"faction_vs"  mapstructure:"faction_vs"`
	ExperienceBonus        string `json:"experience_bonus" mapstructure:"experience_bonus"`
	Timestamp              string `json:"timestamp" mapstructure:"timestamp"`
	ZoneId                 string `json:"zone_id" mapstructure:"zone_id"`
	WorldId                string `json:"world_id" mapstructure:"world_id"`
	EventType              string `json:"event_type" mapstructure:"event_type"`
	InstanceId             string `json:"instance_id" mapstructure:"instance_id"`
	MetagameEventStateName string `json:"metagame_event_state_name" mapstructure:"metagame_event_state_name"`
}

const Character = "character"

type CharacterTimes struct {
	Creation      string `json:"creation" mapstructure:"creation"`
	CreationDate  string `json:"creation_date" mapstructure:"creation_date"`
	LastSave      string `json:"last_save" mapstructure:"last_save"`
	LastSaveDate  string `json:"last_save_date" mapstructure:"last_save_date"`
	LastLogin     string `json:"last_login" mapstructure:"last_login"`
	LastLoginDate string `json:"last_login_date" mapstructure:"last_login_date"`
	LoginCount    string `json:"login_count" mapstructure:"login_count"`
	MinutesPlayed string `json:"minutes_played" mapstructure:"minutes_played"`
}

type CharacterCerts struct {
	EarnedPoints    string `json:"earned_points" mapstructure:"earned_points"`
	GiftedPoints    string `json:"gifted_points" mapstructure:"gifted_points"`
	SpentPoints     string `json:"spent_points" mapstructure:"spent_points"`
	AvailablePoints string `json:"available_points" mapstructure:"available_points"`
	PercentToNext   string `json:"percent_to_next" mapstructure:"percent_to_next"`
}

type CharacterBattleRank struct {
	PercentToNext string `json:"percent_to_next" mapstructure:"percent_to_next"`
	Value         string `json:"value" mapstructure:"value"`
}

type CharacterDailyRibbon struct {
	Count string `json:"count" mapstructure:"count"`
	Time  string `json:"time" mapstructure:"time"`
	Date  string `json:"date" mapstructure:"date"`
}

type CharacterName struct {
	First      string `json:"first" mapstructure:"first"`
	FirstLower string `json:"first_lower" mapstructure:"first_lower"`
}

type CharacterItem struct {
	CharacterId   string               `json:"character_id" mapstructure:"character_id"`
	FactionId     string               `json:"faction_id" mapstructure:"faction_id"`
	HeadId        string               `json:"head_id" mapstructure:"head_id"`
	TitleId       string               `json:"title_id" mapstructure:"title_id"`
	ProfileId     string               `json:"profile_id" mapstructure:"profile_id"`
	PrestigeLevel string               `json:"prestige_level" mapstructure:"prestige_level"`
	Name          CharacterName        `json:"name" mapstructure:"name"`
	Times         CharacterTimes       `json:"times" mapstructure:"times"`
	Certs         CharacterCerts       `json:"certs" mapstructure:"certs"`
	BattleRank    CharacterBattleRank  `json:"battle_rank" mapstructure:"battle_rank"`
	DailyRibbon   CharacterDailyRibbon `json:"daily_ribbon" mapstructure:"daily_ribbon"`
}

const Outfit = "outfit"

type OutfitItem struct {
	OutfitId          string `json:"outfit_id" mapstructure:"outfit_id"`
	Name              string `json:"name" mapstructure:"name"`
	NameLower         string `json:"name_lower" mapstructure:"name_lower"`
	Alias             string `json:"alias" mapstructure:"alias"`
	AliasLower        string `json:"alias_lower" mapstructure:"alias_lower"`
	TimeCreated       string `json:"time_created" mapstructure:"time_created"`
	TimeCreatedDate   string `json:"time_created_date" mapstructure:"time_created_date"`
	LeaderCharacterId string `json:"leader_character_id" mapstructure:"leader_character_id"`
	MemberCount       string `json:"member_count" mapstructure:"member_count"`
	// Resolvable
	Members []CharacterItem `json:"members" mapstructure:"members"`
}

const Map = "map"
const CharactersWorld = "characters_world"
const CharactersOnlineStatus = "characters_online_status"
const CharactersFriend = "characters_friend"
const Leaderboard = "leaderboard"
const CharactersLeaderboard = "characters_leaderboard"
const Event = "event"
const CharactersEvent = "characters_event"
const CharactersEventGrouped = "characters_event_grouped"
const SingleCharacterById = "single_character_by_id"
const CharactersItem = "characters_item"
const World = "world"
