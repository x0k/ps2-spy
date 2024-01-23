package alerts_loader

import (
	"context"
	"log"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type VoidWellLoader struct {
	client *voidwell.Client
}

func NewVoidWell(client *voidwell.Client) *VoidWellLoader {
	return &VoidWellLoader{
		client: client,
	}
}

func (p *VoidWellLoader) Load(ctx context.Context) (loaders.Loaded[ps2.Alerts], error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return loaders.Loaded[ps2.Alerts]{}, err
	}
	// Usually, worlds count is greater than alerts count
	alerts := make(ps2.Alerts, 0, len(states))
	for _, s := range states {
		for _, z := range s.ZoneStates {
			e := z.AlertState.MetagameEvent
			if e.Id == 0 {
				continue
			}
			startedAt, err := time.Parse(time.RFC3339, z.AlertState.Timestamp)
			if err != nil {
				log.Printf("[%s alerts loader] Failed to parse %q: %q", p.client.Endpoint(), z.AlertState.Timestamp, err)
				continue
			}
			duration, err := time.ParseDuration(e.Duration)
			if err != nil {
				log.Printf("[%s alerts loader] Failed to parse %q: %q", p.client.Endpoint(), z.AlertState.Timestamp, err)
				continue
			}
			alert := ps2.Alert{
				WorldId:          ps2.WorldId(s.Id),
				WorldName:        ps2.WorldNames[ps2.WorldId(s.Id)],
				ZoneId:           ps2.ZoneId(e.ZoneId),
				ZoneName:         ps2.ZoneNames[ps2.ZoneId(e.ZoneId)],
				AlertName:        e.Name,
				AlertDescription: e.Description,
				StartedAt:        startedAt,
				Duration:         duration,
			}
			alerts = append(alerts, alert)
		}
	}
	return loaders.LoadedNow(p.client.Endpoint(), alerts), nil
}
