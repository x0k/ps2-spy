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

const Map = "map"
const CharactersWorld = "characters_world"
const CharactersOnlineStatus = "characters_online_status"
const CharactersFriend = "characters_friend"
const Leaderboard = "leaderboard"
const CharactersLeaderboard = "characters_leaderboard"
const Event = "event"
const Characters = "characters_event"
const CharactersEventGrouped = "characters_event_grouped"
const SingleCharacterById = "single_character_by_id"
const CharactersItem = "characters_item"
const World = "world"
