package collections

const WorldEvent = "world_event"

type FacilityWorldEventItem struct {
	FacilityId   string `json:"facility_id" mapstructure:"facility_id"`
	FactionOld   string `json:"faction_old" mapstructure:"faction_old"`
	FactionNew   string `json:"faction_new" mapstructure:"faction_new"`
	DurationHeld string `json:"duration_held" mapstructure:"duration_held"`
	ObjectiveId  string `json:"objective_id" mapstructure:"objective_id"`
	Timestamp    string `json:"timestamp" mapstructure:"timestamp"`
	ZoneId       string `json:"zone_id" mapstructure:"zone_id"`
	WorldId      string `json:"world_id" mapstructure:"world_id"`
	EventType    string `json:"event_type" mapstructure:"event_type"`
	TableType    string `json:"table_type" mapstructure:"table_type"`
	OutfitId     string `json:"outfit_id" mapstructure:"outfit_id"`
}

type MetagameWorldEventItem struct {
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
	// Resolvable
	Outfit OutfitItem `json:"outfit" mapstructure:"outfit"`
	// Resolvable
	WorldId string `json:"world_id" mapstructure:"world_id"`
	// Joinable
	OutfitMemberExtended OutfitMemberExtendedItem `json:"outfit_member_extended" mapstructure:"outfit_member_extended"`
	// Joinable
	CharactersWorld CharactersWorldItem `json:"characters_world" mapstructure:"characters_world"`
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
	// Resolvable
	OutfitMembers []OutfitMemberItem `json:"outfit_members" mapstructure:"outfit_members"`
	// Joinable
	CharactersWorld CharactersWorldItem `json:"characters_world" mapstructure:"characters_world"`
}

const OutfitMember = "outfit_member"

type OutfitMemberItem struct {
	OutfitId        string `json:"outfit_id" mapstructure:"outfit_id"`
	CharacterId     string `json:"character_id" mapstructure:"character_id"`
	MemberSince     string `json:"member_since" mapstructure:"member_since"`
	MemberSinceDate string `json:"member_since_date" mapstructure:"member_since_date"`
	Rank            string `json:"rank" mapstructure:"rank"`
	RankOrdinal     string `json:"rank_ordinal" mapstructure:"rank_ordinal"`
}

const OutfitMemberExtended = "outfit_member_extended"

type OutfitMemberExtendedItem struct {
	CharacterId       string `json:"character_id" mapstructure:"character_id"`
	MemberSince       string `json:"member_since" mapstructure:"member_since"`
	MemberSinceDate   string `json:"member_since_date" mapstructure:"member_since_date"`
	MemberRank        string `json:"member_rank" mapstructure:"member_rank"`
	MemberRankOrdinal string `json:"member_rank_ordinal" mapstructure:"member_rank_ordinal"`
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

const MapRegion = "map_region"

type MapRegionItem struct {
	MapRegionId      string `json:"map_region_id" mapstructure:"map_region_id"`
	ZoneId           string `json:"zone_id" mapstructure:"zone_id"`
	FacilityId       string `json:"facility_id" mapstructure:"facility_id"`
	FacilityName     string `json:"facility_name" mapstructure:"facility_name"`
	FacilityTypeId   string `json:"facility_type_id" mapstructure:"facility_type_id"`
	FacilityType     string `json:"facility_type" mapstructure:"facility_type"`
	LocationX        string `json:"location_x" mapstructure:"location_x"`
	LocationY        string `json:"location_y" mapstructure:"location_y"`
	LocationZ        string `json:"location_z" mapstructure:"location_z"`
	RewardAmount     string `json:"reward_amount" mapstructure:"reward_amount"`
	RewardCurrencyId string `json:"reward_currency_id" mapstructure:"reward_currency_id"`
}

const Map = "map"

type MapItemRowData struct {
	RegionId  string `json:"RegionId" mapstructure:"RegionId"`
	FactionId string `json:"FactionId" mapstructure:"FactionId"`
	// Joinable
	MapRegion MapRegionItem `json:"map_region" mapstructure:"map_region"`
}

type MapItem struct {
	ZoneId  string `json:"ZoneId" mapstructure:"ZoneId"`
	Regions struct {
		IsList string `json:"IsList" mapstructure:"IsList"`
		Row    []struct {
			RowData MapItemRowData `json:"RowData" mapstructure:"RowData"`
		} `json:"Row" mapstructure:"Row"`
	} `json:"Regions" mapstructure:"Regions"`
}

const CharactersWorld = "characters_world"

type CharactersWorldItem struct {
	CharacterId string `json:"character_id" mapstructure:"character_id"`
	WorldId     string `json:"world_id" mapstructure:"world_id"`
}

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
