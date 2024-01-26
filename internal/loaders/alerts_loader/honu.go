package alerts_loader

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type HonuLoader struct {
	client *honu.Client
}

func NewHonu(client *honu.Client) *HonuLoader {
	return &HonuLoader{
		client: client,
	}
}

func (p *HonuLoader) Load(ctx context.Context) (loaders.Loaded[ps2.Alerts], error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return loaders.Loaded[ps2.Alerts]{}, err
	}
	// Usually, worlds count is greater than alerts count
	alerts := make(ps2.Alerts, 0, len(overview))
	for _, w := range overview {
		for _, z := range w.Zones {
			if z.Alert.AlertId != 0 {
				startedAt, err := time.Parse(time.RFC3339, z.Alert.Timestamp)
				if err != nil {
					log.Printf("Failed to parse %q: %q", z.Alert.Timestamp, err)
					continue
				}
				worldId := ps2.WorldId(strconv.Itoa(w.WorldId))
				zoneId := ps2.ZoneId(strconv.Itoa(z.ZoneId))
				alert := ps2.Alert{
					WorldId:          worldId,
					WorldName:        ps2.WorldNames[worldId],
					ZoneId:           zoneId,
					ZoneName:         ps2.ZoneNames[zoneId],
					AlertName:        z.AlertInfo.Name,
					AlertDescription: z.AlertInfo.Description,
					StartedAt:        startedAt,
					Duration:         time.Duration(z.AlertInfo.DurationMinutes) * time.Minute,
					TerritoryControl: ps2.StatsByFactions{
						All:   z.Players.All,
						VS:    z.Players.VS,
						NC:    z.Players.NC,
						TR:    z.Players.TR,
						Other: z.Players.Unknown,
					},
				}
				if alert.WorldName == "" {
					alert.WorldName = fmt.Sprintf("World %s", alert.WorldId)
				}
				if alert.ZoneName == "" {
					alert.ZoneName = fmt.Sprintf("Zone %s", alert.ZoneId)
				}
				alerts = append(alerts, alert)
			}
		}
	}
	return loaders.LoadedNow(p.client.Endpoint(), alerts), nil
}
