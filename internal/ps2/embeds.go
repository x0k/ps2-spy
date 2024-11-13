package ps2

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"time"
)

type Localized struct {
	De string `json:"de"`
	En string `json:"en"`
	Es string `json:"es"`
	Fr string `json:"fr"`
	It string `json:"it"`
	Tr string `json:"tr"`
}

type metagameEventRaw struct {
	MetagameEventId string    `json:"metagame_event_id"`
	Name            Localized `json:"name"`
	Description     Localized `json:"description"`
	Type            string    `json:"type"`
	ExperienceBonus string    `json:"experience_bonus"`
	DurationMinutes string    `json:"duration_minutes"`
}

//go:embed data/metagame_events.json
var metagameEventsFile []byte
var MetagameEventsMap = func() map[MetagameEventId]MetagameEvent {
	var rawInfo map[MetagameEventId]metagameEventRaw
	if err := json.Unmarshal(metagameEventsFile, &rawInfo); err != nil {
		panic(err)
	}
	events := make(map[MetagameEventId]MetagameEvent, len(rawInfo))
	for k, v := range rawInfo {
		d, err := strconv.Atoi(v.DurationMinutes)
		if err != nil {
			panic(err)
		}
		events[MetagameEventId(k)] = MetagameEvent{
			Id:          MetagameEventId(k),
			Name:        v.Name.En,
			Description: v.Description.En,
			Duration:    time.Duration(d) * time.Minute,
		}
	}
	return events
}()
