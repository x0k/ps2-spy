package ps2alerts

type AlertResult struct {
	VS        int `json:"vs"`
	NC        int `json:"nc"`
	TR        int `json:"tr"`
	Cuttoff   int `json:"cuttoff"`
	OutOfPlay int `json:"outOfPlay"`
	// victor // i see only null value
	Draw              bool    `json:"draw"`
	PerBasePercentage float64 `json:"perBasePercentage"`
}

type AlertFeatures struct {
	CaptureHistory bool `json:"captureHistory"`
	Xpm            bool `json:"xpm"`
}

type Alert struct {
	World                   int           `json:"world"`
	CensusInstanceId        int           `json:"censusInstanceId"`
	InstanceId              string        `json:"instanceId"`
	Zone                    int           `json:"zone"`
	TimeStarted             string        `json:"timeStarted"`
	TimeEnded               string        `json:"timeEnded"`
	CensusMetagameEventType int           `json:"censusMetagameEventType"`
	Duration                int           `json:"duration"`
	State                   int           `json:"state"`
	Ps2AlertsEventType      int           `json:"ps2AlertsEventType"`
	Bracket                 int           `json:"bracket"`
	MapVersion              string        `json:"mapVersion"`
	LatticeVersion          string        `json:"latticeVersion"`
	Result                  AlertResult   `json:"result"`
	Features                AlertFeatures `json:"features"`
}
