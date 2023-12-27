package ps2

import (
	"context"
	"sync"
	"time"

	"github.com/x0k/ps2-feed/internal/contextx"
)

type populationProvider interface {
	Name() string
	Population() (Population, error)
}

type Service struct {
	m                  sync.Mutex
	populationProvider populationProvider
	population         Population
	updatedAt          time.Time
}

func NewService(populationProvider populationProvider) *Service {
	return &Service{
		populationProvider: populationProvider,
		updatedAt:          time.Now().Add(-1 * time.Minute),
	}
}

func (s *Service) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Service) PopulationSource() string {
	return s.populationProvider.Name()
}

func (s *Service) isPopulationActual() bool {
	return time.Since(s.updatedAt) < time.Minute
}

func (s *Service) Population(ctx context.Context) (Population, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.isPopulationActual() {
		return s.population, nil
	}
	var err error
	s.population, err = contextx.Go(ctx, s.populationProvider.Population)
	if err != nil {
		return s.population, err
	}
	s.updatedAt = time.Now()
	return s.population, nil
}

func populationByWorldId(p Population, worldId WorldId) (WorldPopulation, error) {
	if p, ok := p.Worlds[worldId]; ok {
		return p, nil
	}
	return WorldPopulation{}, ErrWorldNotFound
}

func (s *Service) PopulationByWorldId(ctx context.Context, worldId WorldId) (WorldPopulation, error) {
	if s.isPopulationActual() {
		return populationByWorldId(s.population, worldId)
	}
	population, err := s.Population(ctx)
	if err != nil {
		return WorldPopulation{}, err
	}
	return populationByWorldId(population, worldId)
}
