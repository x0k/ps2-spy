package fisu

type Response[R any] struct {
	Result R `json:"result"`
}

type WorldPopulation struct {
	WorldId   int `json:"worldId"`
	Timestamp int `json:"timestamp"`
	VS        int `json:"vs"`
	NC        int `json:"nc"`
	TR        int `json:"tr"`
	NS        int `json:"ns"`
	Unknown   int `json:"unknown"`
}

type WorldsPopulation map[string][]WorldPopulation
