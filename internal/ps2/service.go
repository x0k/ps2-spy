package ps2

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type Service struct {
	worldsPopulation *containers.LoadableValue[Loaded[WorldsPopulation]]
	worldPopulation  *containers.KeyedLoadableValues[WorldId, Loaded[DetailedWorldPopulation]]
	alerts           *containers.LoadableValue[Loaded[Alerts]]
}

func NewService(
	worldsPopulationProvider Loader[WorldsPopulation],
	worldPopulationProvider KeyedLoader[WorldId, DetailedWorldPopulation],
	alertsProvider Loader[Alerts],
) *Service {
	return &Service{
		worldsPopulation: containers.NewLoadableValue(worldsPopulationProvider, time.Minute),
		worldPopulation:  containers.NewKeyedLoadableValue(worldPopulationProvider, 10, time.Minute),
		alerts:           containers.NewLoadableValue(alertsProvider, time.Minute),
	}
}

func (s *Service) Start() {
	go s.worldsPopulation.StartExpiration()
	go s.alerts.StartExpiration()
}

func (s *Service) Stop() {
	s.worldsPopulation.StopExpiration()
	s.alerts.StopExpiration()
}

func (s *Service) Population(ctx context.Context) (Loaded[WorldsPopulation], error) {
	return s.worldsPopulation.Load(ctx)
}

func (s *Service) PopulationByWorldId(ctx context.Context, worldId WorldId) (Loaded[DetailedWorldPopulation], error) {
	return s.worldPopulation.Load(ctx, worldId)
}

func (s *Service) Alerts(ctx context.Context) (Loaded[Alerts], error) {
	return s.alerts.Load(ctx)
}

func (s *Service) AlertsByWorldId(ctx context.Context, worldId WorldId) (Loaded[Alerts], error) {
	loaded, err := s.alerts.Load(ctx)
	if err != nil {
		return Loaded[Alerts]{}, err
	}
	worldAlerts := make(Alerts, 0, len(loaded.Value))
	for _, a := range loaded.Value {
		if a.WorldId == worldId {
			worldAlerts = append(worldAlerts, a)
		}
	}
	loaded.Value = worldAlerts
	return loaded, nil
}
