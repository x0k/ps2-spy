package ps2alerts_alerts_loader

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	client *ps2alerts.Client,
) loader.Simple[meta.Loaded[ps2.Alerts]] {
	return func(ctx context.Context) (meta.Loaded[ps2.Alerts], error) {
		ps2alerts, err := client.Alerts(ctx)
		if err != nil {
			return meta.Loaded[ps2.Alerts]{}, err
		}
		alerts := make(ps2.Alerts, 0, len(ps2alerts))
		for _, a := range ps2alerts {
			alertInfo, ok := ps2.MetagameEventsMap[ps2.MetagameEventId(
				strconv.Itoa(a.CensusMetagameEventType),
			)]
			if !ok {
				alertInfo = ps2.MetagameEvent{
					Name:        fmt.Sprintf("Unknown alert (%d)", a.CensusMetagameEventType),
					Description: "This alert is not registered yet",
				}
			}
			startedAt, err := time.Parse(time.RFC3339, a.TimeStarted)
			if err != nil {
				log.Printf("Failed to parse %q: %q", a.TimeStarted, err)
				continue
			}
			worldId := ps2.WorldId(strconv.Itoa(a.World))
			zoneId := ps2.ZoneId(strconv.Itoa(a.Zone))
			alerts = append(alerts, ps2.Alert{
				WorldId:          worldId,
				WorldName:        ps2.WorldNames[worldId],
				ZoneId:           zoneId,
				ZoneName:         ps2.ZoneNames[zoneId],
				AlertName:        alertInfo.Name,
				AlertDescription: alertInfo.Description,
				StartedAt:        startedAt,
				Duration:         time.Duration(a.Duration) * time.Millisecond,
				TerritoryControl: ps2.StatPerFactions{
					All:   100,
					VS:    a.Result.VS,
					NC:    a.Result.NC,
					TR:    a.Result.TR,
					Other: a.Result.OutOfPlay,
				},
			})
		}
		return meta.LoadedNow(client.Endpoint(), alerts), nil
	}

}
