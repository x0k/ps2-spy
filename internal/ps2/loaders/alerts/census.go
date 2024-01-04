package alerts

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/collections"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type CensusLoader struct {
	client *census2.Client
}

func NewCensusLoader(client *census2.Client) *CensusLoader {
	return &CensusLoader{
		client: client,
	}
}

var WorldEventsQuery = census2.NewQueryMustBeValid(census2.GetQuery, census2.Ns_ps2V2, collections.WorldEvent).
	Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
	Where(census2.Cond("world_id").Equals(census2.Str("1,10,13,17,19,24,40,1000,2000"))).
	SetLimit(100)

func (l *CensusLoader) Load(ctx context.Context) (ps2.Loaded[ps2.Alerts], error) {
	events, err := census2.ExecuteAndDecode[collections.WorldEventItem](ctx, l.client, WorldEventsQuery)
	if err != nil {
		return ps2.Loaded[ps2.Alerts]{}, err
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
			log.Printf("[%s] Failed to parse event id %q: %q", l.client.Endpoint(), e.MetagameEventId, err)
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
			log.Printf("[%s] Failed to parse world id %q: %q", l.client.Endpoint(), e.WorldId, err)
			continue
		}
		zoneId, err := strconv.Atoi(e.ZoneId)
		if err != nil {
			log.Printf("[%s] Failed to parse zone id %q: %q", l.client.Endpoint(), e.ZoneId, err)
			continue
		}
		timesamp, err := strconv.ParseInt(e.Timestamp, 10, 64)
		if err != nil {
			log.Printf("[%s] Failed to parse timestamp %q: %q", l.client.Endpoint(), e.Timestamp, err)
			continue
		}
		var duration time.Duration
		if eventInfo, ok := ps2.MetagameEventsMap[eventId]; ok {
			d, err := strconv.Atoi(eventInfo.DurationMinutes)
			if err != nil {
				log.Printf("[%s] Failed to parse duration %q: %q", l.client.Endpoint(), eventInfo.DurationMinutes, err)
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
	return ps2.LoadedNow(l.client.Endpoint(), alerts), nil
}
