package ps2

import (
	_ "embed"
	"encoding/json"
	"strconv"
)

type AlertInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Localized struct {
	De string `json:"de"`
	En string `json:"en"`
	Es string `json:"es"`
	Fr string `json:"fr"`
	It string `json:"it"`
	Tr string `json:"tr"`
}

//go:embed data/alerts.json
var alertsFile []byte
var AlertsMap = func() map[int]AlertInfo {
	var rawInfo map[string]AlertInfo
	if err := json.Unmarshal(alertsFile, &rawInfo); err != nil {
		panic(err)
	}
	alerts := make(map[int]AlertInfo, len(rawInfo))
	for k, v := range rawInfo {
		id, err := strconv.Atoi(k)
		if err != nil {
			panic(err)
		}
		alerts[id] = v
	}
	return alerts
}()

type MetagameEvent struct {
	MetagameEventId string    `json:"metagame_event_id"`
	Name            Localized `json:"name"`
	Description     Localized `json:"description"`
	Type            string    `json:"type"`
	ExperienceBonus string    `json:"experience_bonus"`
	DurationMinutes string    `json:"duration_minutes"`
}

//go:embed data/metagame_events.json
var metagameEventsFile []byte
var MetagameEventsMap = func() map[int]MetagameEvent {
	var rawInfo map[string]MetagameEvent
	if err := json.Unmarshal(metagameEventsFile, &rawInfo); err != nil {
		panic(err)
	}
	events := make(map[int]MetagameEvent, len(rawInfo))
	for k, v := range rawInfo {
		id, err := strconv.Atoi(k)
		if err != nil {
			panic(err)
		}
		events[id] = v
	}
	return events
}()
