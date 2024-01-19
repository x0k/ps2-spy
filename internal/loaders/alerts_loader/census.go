package alerts_loader

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type CensusLoader struct {
	client   *census2.Client
	pcUrl    string
	ps4euUrl string
	ps4usUrl string
}

func NewCensus(client *census2.Client) *CensusLoader {
	return &CensusLoader{
		client: client,
		pcUrl: client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2_v2_NS, collections.WorldEvent).
			Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
			Where(census2.Cond("world_id").Equals(census2.Str("1,10,13,17,19,24,40"))).
			SetLimit(210)),
		ps4euUrl: client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2ps4eu_v2_NS, collections.WorldEvent).
			Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
			Where(census2.Cond("world_id").Equals(census2.Str("2000"))).
			SetLimit(30)),
		ps4usUrl: client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2ps4us_v2_NS, collections.WorldEvent).
			Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
			Where(census2.Cond("world_id").Equals(census2.Str("1000"))).
			SetLimit(30)),
	}
}

func (l *CensusLoader) load(ctx context.Context, url string) (ps2.Alerts, error) {
	const op = "loaders.alerts_loader.CensusLoader.load"
	log := infra.OpLogger(ctx, op).With(slog.String("census_endpoint", l.client.Endpoint()))
	events, err := census2.ExecutePreparedAndDecode[collections.WorldEventItem](ctx, l.client, collections.WorldEvent, url)
	if err != nil {
		return ps2.Alerts{}, err
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
	return alerts, nil
}

func (l *CensusLoader) Load(ctx context.Context) (loaders.Loaded[ps2.Alerts], error) {
	errors := make([]string, 0, 3)
	pcAlerts, err := l.load(ctx, l.pcUrl)
	if err != nil {
		errors = append(errors, err.Error())
	}
	ps4euAlerts, err := l.load(ctx, l.ps4euUrl)
	if err != nil {
		errors = append(errors, err.Error())
	}
	ps4usAlerts, err := l.load(ctx, l.ps4usUrl)
	if err != nil {
		errors = append(errors, err.Error())
	}
	if len(errors) > 0 {
		return loaders.Loaded[ps2.Alerts]{}, fmt.Errorf("failed to load alerts: %s", strings.Join(errors, ", "))
	}
	alerts := make(ps2.Alerts, 0, len(pcAlerts)+len(ps4euAlerts)+len(ps4usAlerts))
	alerts = append(alerts, pcAlerts...)
	alerts = append(alerts, ps4euAlerts...)
	alerts = append(alerts, ps4usAlerts...)
	return loaders.LoadedNow(l.client.Endpoint(), alerts), nil
}
