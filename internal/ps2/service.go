package ps2

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
)

type provider[T any] interface {
	Name() string
	Load(ctx context.Context) (T, error)
}

type Service struct {
	populationProvider provider[Population]
	population         *containers.LoadableValue[Population]
	alertsProvider     provider[Alerts]
	alerts             *containers.LoadableValue[Alerts]
}

func NewService(populationProvider provider[Population], alertsProvider provider[Alerts]) *Service {
	return &Service{
		populationProvider: populationProvider,
		population:         containers.NewLoadableValue[Population](populationProvider, time.Minute),
		alertsProvider:     alertsProvider,
		alerts:             containers.NewLoadableValue[Alerts](alertsProvider, time.Minute),
	}
}

func (s *Service) Start() {
	go s.population.StartExpiration()
	go s.alerts.StartExpiration()
}

func (s *Service) Stop() {
	s.population.StopExpiration()
	s.alerts.StopExpiration()
}

func (s *Service) PopulationUpdatedAt() time.Time {
	return s.population.UpdatedAt()
}

func (s *Service) PopulationSource() string {
	return s.populationProvider.Name()
}

func (s *Service) Population(ctx context.Context) (Population, error) {
	return s.population.Load(ctx)
}

func (s *Service) PopulationByWorldId(ctx context.Context, worldId WorldId) (WorldPopulation, error) {
	population, err := s.population.Load(ctx)
	if err != nil {
		return WorldPopulation{}, err
	}
	if p, ok := population.Worlds[worldId]; ok {
		return p, nil
	}
	return WorldPopulation{}, ErrWorldNotFound
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
