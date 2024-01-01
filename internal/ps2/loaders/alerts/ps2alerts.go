package alerts

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type PS2AlertsLoader struct {
	client *ps2alerts.Client
}

func NewPS2AlertsLoader(client *ps2alerts.Client) *PS2AlertsLoader {
	return &PS2AlertsLoader{
		client: client,
	}
}

func (p *PS2AlertsLoader) Load(ctx context.Context) (ps2.Loaded[ps2.Alerts], error) {
	ps2alerts, err := p.client.Alerts(ctx)
	if err != nil {
		return ps2.Loaded[ps2.Alerts]{}, err
	}
	alerts := make(ps2.Alerts, 0, len(ps2alerts))
	for _, a := range ps2alerts {
		alertInfo, ok := ps2.AlertsMap[a.CensusMetagameEventType]
		if !ok {
			alertInfo = ps2.AlertInfo{
				Name:        fmt.Sprintf("Unknown alert (%d)", a.CensusMetagameEventType),
				Description: "This alert is not registered yet",
			}
		}
		startedAt, err := time.Parse(time.RFC3339, a.TimeStarted)
		if err != nil {
			log.Printf("Failed to parse %q: %q", a.TimeStarted, err)
			continue
		}
		alerts = append(alerts, ps2.Alert{
			WorldId:          ps2.WorldId(a.World),
			WorldName:        ps2.WorldNames[ps2.WorldId(a.World)],
			ZoneId:           ps2.ZoneId(a.Zone),
			ZoneName:         ps2.ZoneNames[ps2.ZoneId(a.Zone)],
			AlertName:        alertInfo.Name,
			AlertDescription: alertInfo.Description,
			StartedAt:        startedAt,
			Duration:         time.Duration(a.Duration) * time.Millisecond,
		})
	}
	return ps2.LoadedNow(p.client.Endpoint(), alerts), nil
}
