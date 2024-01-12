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
