package ps2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/x0k/ps2-spy/internal/honu"
)

type HonuAlertsLoader struct {
	client *honu.Client
}

func NewHonuAlertsLoader(client *honu.Client) *HonuAlertsLoader {
	return &HonuAlertsLoader{
		client: client,
	}
}

func (p *HonuAlertsLoader) Load(ctx context.Context) (Alerts, error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return Alerts{}, err
	}
	// Usually, worlds count is greater than alerts count
	alerts := make(Alerts, 0, len(overview))
	for _, w := range overview {
		for _, z := range w.Zones {
			if z.Alert.AlertId != 0 {
				startedAt, err := time.Parse(time.RFC3339, z.Alert.Timestamp)
				if err != nil {
					log.Printf("Failed to parse %q: %q", z.Alert.Timestamp, err)
					continue
				}
				alert := Alert{
					WorldId:          WorldId(w.WorldId),
					WorldName:        WorldNames[WorldId(w.WorldId)],
					ZoneId:           ZoneId(z.ZoneId),
					ZoneName:         ZoneNames[ZoneId(z.ZoneId)],
					AlertName:        z.AlertInfo.Name,
					AlertDescription: z.AlertInfo.Description,
					StartedAt:        startedAt,
					Duration:         time.Duration(z.AlertInfo.DurationMinutes) * time.Minute,
					TerritoryControl: StatsByFactions{
						All:   z.Players.All,
						VS:    z.Players.VS,
						NC:    z.Players.NC,
						TR:    z.Players.TR,
						Other: z.Players.Unknown,
					},
				}
				if alert.WorldName == "" {
					alert.WorldName = fmt.Sprintf("World %d", alert.WorldId)
				}
				if alert.ZoneName == "" {
					alert.ZoneName = fmt.Sprintf("Zone %d", alert.ZoneId)
				}
				alerts = append(alerts, alert)
			}
		}
	}
	return alerts, nil
}
