package sanctuary

const WorldPopulationCollection = "world_population"

type Factions struct {
	NC  int `json:"NC"`
	TR  int `json:"TR"`
	VS  int `json:"VS"`
	NSO int `json:"NSO"`
}

type WorldPopulation struct {
	WorldId    int      `json:"world_id"`
	Timestamp  int      `json:"timestamp"`
	Total      int      `json:"total"`
	Population Factions `json:"population"`
	EventName  string   `json:"event_name"`
}
