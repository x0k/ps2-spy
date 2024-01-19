package alerts_loader

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type CensusLoader struct {
	client *census2.Client
}

func NewCensus(client *census2.Client) *CensusLoader {
	return &CensusLoader{
		client: client,
	}
}

var PcWorldEventsQuery = census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2_v2_NS, collections.WorldEvent).
	Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
	Where(census2.Cond("world_id").Equals(census2.Str("1,10,13,17,19,24,40,1000,2000"))).
	SetLimit(100)

func (l *CensusLoader) Load(ctx context.Context) (loaders.Loaded[ps2.Alerts], error) {
	const op = "loaders.alerts_loader.CensusLoader.Load"
	log := infra.OpLogger(ctx, op).With(slog.String("census_endpoint", l.client.Endpoint()))
	events, err := census2.ExecuteAndDecode[collections.WorldEventItem](ctx, l.client, PcWorldEventsQuery)
	if err != nil {
		return loaders.Loaded[ps2.Alerts]{}, err
	}
	actualEvents := make(map[string]collections.WorldEventItem, len(events))
	for i := len(events) - 1; i >= 0; i-- {
		e := events[i]
		if e.MetagameEventStateName == "started" {
			actualEvents[e.InstanceId] = e
		} else {
			delete(actualEvents, e.InstanceId)
		}
	}
	alerts := make(ps2.Alerts, 0, len(actualEvents))
	for _, e := range actualEvents {
		eventId, err := strconv.Atoi(e.MetagameEventId)
		if err != nil {
			log.Error("failed to parse event id", slog.String("event_id", e.MetagameEventId), sl.Err(err))
			continue
		}
		alertInfo, ok := ps2.AlertsMap[eventId]
		if !ok {
			alertInfo = ps2.AlertInfo{
				Name:        fmt.Sprintf("Unknown alert (%d)", eventId),
				Description: "This alert is not registered yet",
			}
		}
		worldId, err := strconv.Atoi(e.WorldId)
		if err != nil {
			log.Error("failed to parse world id", slog.String("world_id", e.WorldId), sl.Err(err))
			continue
		}
		zoneId, err := strconv.Atoi(e.ZoneId)
		if err != nil {
			log.Error("failed to parse zone id", slog.String("zone_id", e.ZoneId), sl.Err(err))
			continue
		}
		timesamp, err := strconv.ParseInt(e.Timestamp, 10, 64)
		if err != nil {
			log.Error("failed to parse timestamp", slog.String("timestamp", e.Timestamp), sl.Err(err))
			continue
		}
		var duration time.Duration
		if eventInfo, ok := ps2.MetagameEventsMap[eventId]; ok {
			d, err := strconv.Atoi(eventInfo.DurationMinutes)
			if err != nil {
				log.Error("failed to parse duration", slog.String("duration", eventInfo.DurationMinutes), sl.Err(err))
			} else {
				duration = time.Duration(d) * time.Minute
			}
		}
		startedAt := time.Unix(timesamp, 0)
		nc, _ := strconv.ParseFloat(e.FactionNC, 64)
		vs, _ := strconv.ParseFloat(e.FactionVS, 64)
		tr, _ := strconv.ParseFloat(e.FactionTR, 64)
		alert := ps2.Alert{
			WorldId:          ps2.WorldId(worldId),
			WorldName:        ps2.WorldNames[ps2.WorldId(worldId)],
			ZoneId:           ps2.ZoneId(zoneId),
			ZoneName:         ps2.ZoneNames[ps2.ZoneId(zoneId)],
			AlertName:        alertInfo.Name,
			AlertDescription: alertInfo.Description,
			StartedAt:        startedAt,
			Duration:         duration,
			TerritoryControl: ps2.StatsByFactions{
				All: 100,
				NC:  int(nc),
				VS:  int(vs),
				TR:  int(tr),
			},
		}
		alerts = append(alerts, alert)
	}
	return loaders.LoadedNow(l.client.Endpoint(), alerts), nil
}
