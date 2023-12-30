package ps2

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
)

type loader[T any] interface {
	Name() string
	Load(ctx context.Context) (T, error)
}

type keyedLoader[K comparable, T any] interface {
	Name() string
	Load(ctx context.Context, key K) (T, error)
}

type Service struct {
	worldsPopulationProvider loader[WorldsPopulation]
	worldsPopulation         *containers.LoadableValue[WorldsPopulation]
	worldPopulationProvider  keyedLoader[WorldId, WorldPopulation]
	worldPopulation          *containers.KeyedLoadableValues[WorldId, WorldPopulation]
	alertsProvider           loader[Alerts]
	alerts                   *containers.LoadableValue[Alerts]
}

func NewService(
	worldsPopulationProvider loader[WorldsPopulation],
	worldPopulationProvider keyedLoader[WorldId, WorldPopulation],
	alertsProvider loader[Alerts],
) *Service {
	return &Service{
		worldsPopulationProvider: worldsPopulationProvider,
		worldsPopulation:         containers.NewLoadableValue(worldsPopulationProvider, time.Minute),
		worldPopulationProvider:  worldPopulationProvider,
		worldPopulation:          containers.NewKeyedLoadableValue(worldPopulationProvider, 10, time.Minute),
		alertsProvider:           alertsProvider,
		alerts:                   containers.NewLoadableValue(alertsProvider, time.Minute),
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

func (s *Service) PopulationUpdatedAt() time.Time {
	return s.worldsPopulation.UpdatedAt()
}

func (s *Service) PopulationSource() string {
	return s.worldsPopulationProvider.Name()
}

func (s *Service) Population(ctx context.Context) (WorldsPopulation, error) {
	return s.worldsPopulation.Load(ctx)
}

func (s *Service) PopulationByWorldId(ctx context.Context, worldId WorldId) (WorldPopulation, error) {
	return s.worldPopulation.Load(ctx, worldId)
}

func (s *Service) AlertsSource() string {
	return s.alertsProvider.Name()
}

func (s *Service) AlertsUpdatedAt() time.Time {
	return s.alerts.UpdatedAt()
}

func (s *Service) Alerts(ctx context.Context) (Alerts, error) {
	return s.alerts.Load(ctx)
}

func (s *Service) AlertsByWorldId(ctx context.Context, worldId WorldId) (Alerts, error) {
	alerts, err := s.alerts.Load(ctx)
	if err != nil {
		return Alerts{}, err
	}
	worldAlerts := make(Alerts, 0, len(alerts))
	for _, a := range alerts {
		if a.WorldId == worldId {
			worldAlerts = append(worldAlerts, a)
		}
	}
	return worldAlerts, nil
}
