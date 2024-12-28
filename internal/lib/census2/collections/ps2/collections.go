package ps2_collections

const WorldEvent = "world_event"

type FacilityWorldEventItem struct {
	FacilityId   string `json:"facility_id"`
	FactionOld   string `json:"faction_old"`
	FactionNew   string `json:"faction_new"`
	DurationHeld string `json:"duration_held"`
	ObjectiveId  string `json:"objective_id"`
	Timestamp    string `json:"timestamp"`
	ZoneId       string `json:"zone_id"`
	WorldId      string `json:"world_id"`
	EventType    string `json:"event_type"`
	TableType    string `json:"table_type"`
	OutfitId     string `json:"outfit_id"`
}

type MetagameWorldEventItem struct {
	MetagameEventId        string `json:"metagame_event_id"`
	MetagameEventState     string `json:"metagame_event_state"`
	FactionNC              string `json:"faction_nc"`
	FactionTR              string `json:"faction_tr"`
	FactionVS              string `json:"faction_vs" `
	ExperienceBonus        string `json:"experience_bonus"`
	Timestamp              string `json:"timestamp"`
	ZoneId                 string `json:"zone_id"`
	WorldId                string `json:"world_id"`
	EventType              string `json:"event_type"`
	InstanceId             string `json:"instance_id"`
	MetagameEventStateName string `json:"metagame_event_state_name"`
}

const Character = "character"

type CharacterTimes struct {
	Creation      string `json:"creation"`
	CreationDate  string `json:"creation_date"`
	LastSave      string `json:"last_save"`
	LastSaveDate  string `json:"last_save_date"`
	LastLogin     string `json:"last_login"`
	LastLoginDate string `json:"last_login_date"`
	LoginCount    string `json:"login_count"`
	MinutesPlayed string `json:"minutes_played"`
}

type CharacterCerts struct {
	EarnedPoints    string `json:"earned_points"`
	GiftedPoints    string `json:"gifted_points"`
	SpentPoints     string `json:"spent_points"`
	AvailablePoints string `json:"available_points"`
	PercentToNext   string `json:"percent_to_next"`
}

type CharacterBattleRank struct {
	PercentToNext string `json:"percent_to_next"`
	Value         string `json:"value"`
}

type CharacterDailyRibbon struct {
	Count string `json:"count"`
	Time  string `json:"time"`
	Date  string `json:"date"`
}

type CharacterName struct {
	First      string `json:"first"`
	FirstLower string `json:"first_lower"`
}

type CharacterItem struct {
	CharacterId   string               `json:"character_id"`
	FactionId     string               `json:"faction_id"`
	HeadId        string               `json:"head_id"`
	TitleId       string               `json:"title_id"`
	ProfileId     string               `json:"profile_id"`
	PrestigeLevel string               `json:"prestige_level"`
	Name          CharacterName        `json:"name"`
	Times         CharacterTimes       `json:"times"`
	Certs         CharacterCerts       `json:"certs"`
	BattleRank    CharacterBattleRank  `json:"battle_rank"`
	DailyRibbon   CharacterDailyRibbon `json:"daily_ribbon"`
	// Resolvable
	Outfit OutfitItem `json:"outfit"`
	// Resolvable
	WorldId string `json:"world_id"`
	// Joinable
	OutfitMemberExtended OutfitMemberExtendedItem `json:"outfit_member_extended"`
	// Joinable
	CharactersWorld CharactersWorldItem `json:"characters_world"`
}

const Outfit = "outfit"

type OutfitItem struct {
	OutfitId          string `json:"outfit_id"`
	Name              string `json:"name"`
	NameLower         string `json:"name_lower"`
	Alias             string `json:"alias"`
	AliasLower        string `json:"alias_lower"`
	TimeCreated       string `json:"time_created"`
	TimeCreatedDate   string `json:"time_created_date"`
	LeaderCharacterId string `json:"leader_character_id"`
	MemberCount       string `json:"member_count"`
	// Resolvable
	Members []CharacterItem `json:"members"`
	// Resolvable
	OutfitMembers []OutfitMemberItem `json:"outfit_members"`
	// Joinable
	CharactersWorld CharactersWorldItem `json:"characters_world"`
}

const OutfitMember = "outfit_member"

type OutfitMemberItem struct {
	OutfitId        string `json:"outfit_id"`
	CharacterId     string `json:"character_id"`
	MemberSince     string `json:"member_since"`
	MemberSinceDate string `json:"member_since_date"`
	Rank            string `json:"rank"`
	RankOrdinal     string `json:"rank_ordinal"`
}

const OutfitMemberExtended = "outfit_member_extended"

type OutfitMemberExtendedItem struct {
	CharacterId       string `json:"character_id"`
	MemberSince       string `json:"member_since"`
	MemberSinceDate   string `json:"member_since_date"`
	MemberRank        string `json:"member_rank"`
	MemberRankOrdinal string `json:"member_rank_ordinal"`
	OutfitId          string `json:"outfit_id"`
	Name              string `json:"name"`
	NameLower         string `json:"name_lower"`
	Alias             string `json:"alias"`
	AliasLower        string `json:"alias_lower"`
	TimeCreated       string `json:"time_created"`
	TimeCreatedDate   string `json:"time_created_date"`
	LeaderCharacterId string `json:"leader_character_id"`
	MemberCount       string `json:"member_count"`
}

const MapRegion = "map_region"

type MapRegionItem struct {
	MapRegionId      string `json:"map_region_id"`
	ZoneId           string `json:"zone_id"`
	FacilityId       string `json:"facility_id"`
	FacilityName     string `json:"facility_name"`
	FacilityTypeId   string `json:"facility_type_id"`
	FacilityType     string `json:"facility_type"`
	LocationX        string `json:"location_x"`
	LocationY        string `json:"location_y"`
	LocationZ        string `json:"location_z"`
	RewardAmount     string `json:"reward_amount"`
	RewardCurrencyId string `json:"reward_currency_id"`
}

const Map = "map"

type MapItemRowData struct {
	RegionId  string `json:"RegionId"`
	FactionId string `json:"FactionId"`
	// Joinable
	MapRegion MapRegionItem `json:"map_region"`
}

type MapItem struct {
	ZoneId  string `json:"ZoneId"`
	Regions struct {
		IsList string `json:"IsList"`
		Row    []struct {
			RowData MapItemRowData `json:"RowData"`
		} `json:"Row"`
	} `json:"Regions"`
}

const CharactersWorld = "characters_world"

type CharactersWorldItem struct {
	CharacterId string `json:"character_id"`
	WorldId     string `json:"world_id"`
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
