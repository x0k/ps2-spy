package census_alerts_loader

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	log *logger.Logger,
	client *census2.Client,
) loader.Simple[meta.Loaded[ps2.Alerts]] {
	pcUrl := client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.WorldEvent).
		Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
		Where(census2.Cond("world_id").Equals(census2.Str("1,10,13,17,19,24,40"))).
		SetLimit(210))
	ps4euUrl := client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2ps4eu_v2_NS, ps2_collections.WorldEvent).
		Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
		Where(census2.Cond("world_id").Equals(census2.Str("2000"))).
		SetLimit(30))
	ps4usUrl := client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2ps4us_v2_NS, ps2_collections.WorldEvent).
		Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
		Where(census2.Cond("world_id").Equals(census2.Str("1000"))).
		SetLimit(30))
	load := func(ctx context.Context, url string) (ps2.Alerts, error) {
		events, err := census2.ExecutePreparedAndDecode[ps2_collections.MetagameWorldEventItem](ctx, client, ps2_collections.WorldEvent, url)
		if err != nil {
			return ps2.Alerts{}, err
		}
		actualEvents := make(map[string]ps2_collections.MetagameWorldEventItem, len(events))
		for i := len(events) - 1; i >= 0; i-- {
			e := events[i]
			if e.MetagameEventStateName == ps2.StartedMetagameEventStateName {
				actualEvents[e.InstanceId] = e
			} else {
				delete(actualEvents, e.InstanceId)
			}
		}
		alerts := make(ps2.Alerts, 0, len(actualEvents))
		for _, e := range actualEvents {
			eventId := ps2.MetagameEventId(e.MetagameEventId)
			worldId := ps2.WorldId(e.WorldId)
			zoneId := ps2.ZoneId(e.ZoneId)
			timesamp, err := strconv.ParseInt(e.Timestamp, 10, 64)
			if err != nil {
				log.Error(ctx, "failed to parse timestamp", slog.String("timestamp", e.Timestamp), sl.Err(err))
				continue
			}
			eventInfo, ok := ps2.MetagameEventsMap[eventId]
			if !ok {
				eventInfo = ps2.MetagameEvent{
					Name:        fmt.Sprintf("Unknown alert (%s)", eventId),
					Description: "This event is not registered yet",
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
				AlertName:        eventInfo.Name,
				AlertDescription: eventInfo.Description,
				StartedAt:        startedAt,
				Duration:         eventInfo.Duration,
				TerritoryControl: ps2.StatPerFactions{
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
	return func(ctx context.Context) (meta.Loaded[ps2.Alerts], error) {
		errors := make([]string, 0, 3)
		pcAlerts, err := load(ctx, pcUrl)
		if err != nil {
			errors = append(errors, err.Error())
		}
		ps4euAlerts, err := load(ctx, ps4euUrl)
		if err != nil {
			errors = append(errors, err.Error())
		}
		ps4usAlerts, err := load(ctx, ps4usUrl)
		if err != nil {
			errors = append(errors, err.Error())
		}
		if len(errors) > 0 {
			return meta.Loaded[ps2.Alerts]{}, fmt.Errorf("failed to load alerts: %s", strings.Join(errors, ", "))
		}
		alerts := make(ps2.Alerts, 0, len(pcAlerts)+len(ps4euAlerts)+len(ps4usAlerts))
		alerts = append(alerts, pcAlerts...)
		alerts = append(alerts, ps4euAlerts...)
		alerts = append(alerts, ps4usAlerts...)
		return meta.LoadedNow(client.Endpoint(), alerts), nil
	}
}
