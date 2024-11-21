package saerro

type GraphqlResponse[T any] struct {
	Data T `json:"data"`
}

type Factions struct {
	Total int `json:"total"`
	NC    int `json:"nc"`
	VS    int `json:"vs"`
	TR    int `json:"tr"`
	NS    int `json:"ns"`
}

type ZonePopulation struct {
	Id         int      `json:"id"`
	Name       string   `json:"name"`
	Population Factions `json:"population"`
}

type AllZonesPopulation struct {
	All []ZonePopulation
}

type WorldPopulation struct {
	Id    int                `json:"id"`
	Name  string             `json:"name"`
	Zones AllZonesPopulation `json:"zones"`
}

type AllWorldsPopulation struct {
	AllWorlds []WorldPopulation `json:"allWorlds"`
}
