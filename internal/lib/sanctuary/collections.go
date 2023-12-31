package sanctuary

const CensusJSONField = "c:censusJSON"

const WorldPopulationCollection = "world_population"

type Factions struct {
	NC  int `json:"NC" mapstructure:"NC"`
	TR  int `json:"TR" mapstructure:"TR"`
	VS  int `json:"VS" mapstructure:"VS"`
	NSO int `json:"NSO" mapstructure:"NSO"`
}

type WorldPopulation struct {
	WorldId    int      `json:"world_id" mapstructure:"world_id"`
	Timestamp  int      `json:"timestamp" mapstructure:"timestamp"`
	Total      int      `json:"total" mapstructure:"total"`
	Population Factions `json:"population" mapstructure:"population"`
	EventName  string   `json:"event_name" mapstructure:"event_name"`
}
