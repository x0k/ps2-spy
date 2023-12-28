package ps2

import (
	"context"
	"time"

	"github.com/x0k/ps2-feed/internal/cache"
)

type populationProvider interface {
	Name() string
	Population(ctx context.Context) (Population, error)
}

type alertsProvider interface {
	Name() string
	Alerts(ctx context.Context) (Alerts, error)
}

type Service struct {
	populationProvider  populationProvider
	population          *cache.ExpiableValue[Population]
	populationUpdatedAt time.Time
	alertsProvider      alertsProvider
	alerts              *cache.ExpiableValue[Alerts]
	alertsUpdatedAt     time.Time
}

func NewService(populationProvider populationProvider, alertsProvider alertsProvider) *Service {
	return &Service{
		populationProvider: populationProvider,
		population:         cache.NewExpiableValue[Population](time.Minute),
		alertsProvider:     alertsProvider,
		alerts:             cache.NewExpiableValue[Alerts](time.Minute),
	}
}

func (s *Service) Stop() {
	s.population.Stop()
	s.alerts.Stop()
}

func (s *Service) PopulationUpdatedAt() time.Time {
	return s.populationUpdatedAt
}

func (s *Service) PopulationSource() string {
	return s.populationProvider.Name()
}

func (s *Service) loadPopulation(ctx context.Context) (Population, error) {
	return s.population.Load(func() (Population, error) {
		population, err := s.populationProvider.Population(ctx)
		if err != nil {
			return Population{}, err
		}
		s.populationUpdatedAt = time.Now()
		return population, nil
	})
}

func (s *Service) Population(ctx context.Context) (Population, error) {
	return s.loadPopulation(ctx)
}

func (s *Service) PopulationByWorldId(ctx context.Context, worldId WorldId) (WorldPopulation, error) {
	population, err := s.loadPopulation(ctx)
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
	return s.alertsUpdatedAt
}

func (s *Service) loadAlerts(ctx context.Context) (Alerts, error) {
	return s.alerts.Load(func() (Alerts, error) {
		alerts, err := s.alertsProvider.Alerts(ctx)
		if err != nil {
			return Alerts{}, err
		}
		s.alertsUpdatedAt = time.Now()
		return alerts, nil
	})
}

func (s *Service) Alerts(ctx context.Context) (Alerts, error) {
	return s.loadAlerts(ctx)
}

func (s *Service) AlertsByWorldId(ctx context.Context, worldId WorldId) (Alerts, error) {
	alerts, err := s.loadAlerts(ctx)
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
