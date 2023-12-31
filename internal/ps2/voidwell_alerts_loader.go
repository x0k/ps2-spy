package ps2

import (
	"context"
	"log"
	"time"

	"github.com/x0k/ps2-spy/internal/voidwell"
)

type VoidWellAlertsLoader struct {
	client *voidwell.Client
}

func NewVoidWellAlertsLoader(client *voidwell.Client) *VoidWellAlertsLoader {
	return &VoidWellAlertsLoader{
		client: client,
	}
}

func (p *VoidWellAlertsLoader) Load(ctx context.Context) (Alerts, error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return Alerts{}, err
	}
	// Usually, worlds count is greater than alerts count
	alerts := make(Alerts, 0, len(states))
	for _, s := range states {
		for _, z := range s.ZoneStates {
			e := z.AlertState.MetagameEvent
			if e.Id == 0 {
				continue
			}
			startedAt, err := time.Parse(time.RFC3339, z.AlertState.Timestamp)
			if err != nil {
				log.Printf("Failed to parse %q: %q", z.AlertState.Timestamp, err)
				continue
			}
			duration, err := time.ParseDuration(e.Duration)
			if err != nil {
				log.Printf("Failed to parse %q: %q", z.AlertState.Timestamp, err)
				continue
			}
			alert := Alert{
				WorldId:          WorldId(s.Id),
				WorldName:        WorldNames[WorldId(s.Id)],
				ZoneId:           ZoneId(e.ZoneId),
				ZoneName:         ZoneNames[ZoneId(e.ZoneId)],
				AlertName:        e.Name,
				AlertDescription: e.Description,
				StartedAt:        startedAt,
				Duration:         duration,
			}
			alerts = append(alerts, alert)
		}
	}
	return alerts, nil
}
