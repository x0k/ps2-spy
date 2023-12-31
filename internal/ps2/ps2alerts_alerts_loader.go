package ps2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/x0k/ps2-spy/internal/ps2alerts"
)

type PS2AlertsAlertsLoader struct {
	client *ps2alerts.Client
}

func NewPS2AlertsAlertsLoader(client *ps2alerts.Client) *PS2AlertsAlertsLoader {
	return &PS2AlertsAlertsLoader{
		client: client,
	}
}

func (p *PS2AlertsAlertsLoader) Load(ctx context.Context) (Alerts, error) {
	ps2alerts, err := p.client.Alerts(ctx)
	if err != nil {
		return Alerts{}, err
	}
	alerts := make(Alerts, 0, len(ps2alerts))
	for _, a := range ps2alerts {
		alertInfo, ok := alertsMap[a.CensusMetagameEventType]
		if !ok {
			alertInfo = AlertInfo{
				Name:        fmt.Sprintf("Unknown alert (%d)", a.CensusMetagameEventType),
				Description: "This alert is not registered yet",
			}
		}
		startedAt, err := time.Parse(time.RFC3339, a.TimeStarted)
		if err != nil {
			log.Printf("Failed to parse %q: %q", a.TimeStarted, err)
			continue
		}
		alerts = append(alerts, Alert{
			WorldId:          WorldId(a.World),
			WorldName:        WorldNames[WorldId(a.World)],
			ZoneId:           ZoneId(a.Zone),
			ZoneName:         ZoneNames[ZoneId(a.Zone)],
			AlertName:        alertInfo.Name,
			AlertDescription: alertInfo.Description,
			StartedAt:        startedAt,
			Duration:         time.Duration(a.Duration) * time.Millisecond,
		})
	}
	return alerts, nil
}
