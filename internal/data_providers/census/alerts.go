package census_data_provider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (p *DataProvider) alerts(ctx context.Context, url string) (ps2.Alerts, error) {
	events, err := census2_adapters.RetryableExecutePreparedAndDecode[ps2_collections.MetagameWorldEventItem](
		ctx, p.log, p.client, ps2_collections.WorldEvent, url,
	)
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
			p.log.Error(ctx, "failed to parse timestamp", slog.String("timestamp", e.Timestamp), sl.Err(err))
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

func (p *DataProvider) Alerts(ctx context.Context) (meta.Loaded[ps2.Alerts], error) {
	errs := make([]error, 0, 3)
	pcAlerts, err := p.alerts(ctx, p.pcUrl)
	if err != nil {
		errs = append(errs, err)
	}
	ps4euAlerts, err := p.alerts(ctx, p.ps4euUrl)
	if err != nil {
		errs = append(errs, err)
	}
	ps4usAlerts, err := p.alerts(ctx, p.ps4usUrl)
	if err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return meta.Loaded[ps2.Alerts]{}, fmt.Errorf("failed to load alerts: %w", errors.Join(errs...))
	}
	alerts := make(ps2.Alerts, 0, len(pcAlerts)+len(ps4euAlerts)+len(ps4usAlerts))
	alerts = append(alerts, pcAlerts...)
	alerts = append(alerts, ps4euAlerts...)
	alerts = append(alerts, ps4usAlerts...)
	return meta.LoadedNow(p.client.Endpoint(), alerts), nil
}
