package ps2

import (
	"context"
	"sync"

	"github.com/x0k/ps2-feed/internal/contextx"
)

type populationProvider interface {
	Population() (Population, error)
}

type Service struct {
	m                  sync.Mutex
	populationProvider populationProvider
	population         Population
}

func NewService(populationProvider populationProvider) *Service {
	return &Service{
		populationProvider: populationProvider,
	}
}

func (s *Service) Population(ctx context.Context) (Population, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.population != nil {
		return s.population, nil
	}
	var err error
	s.population, err = contextx.Go(ctx, s.populationProvider.Population)
	if err != nil {
		return nil, err
	}
	return s.population, nil
}

func populationByWorldId(pop Population, worldId WorldId) (WorldPopulation, error) {
	if p, ok := pop[worldId]; ok {
		return p, nil
	}
	return WorldPopulation{}, ErrWorldNotFound
}

func (s *Service) PopulationByWorldId(ctx context.Context, worldId WorldId) (WorldPopulation, error) {
	if s.population != nil {
		return populationByWorldId(s.population, worldId)
	}
	population, err := s.Population(ctx)
	if err != nil {
		return WorldPopulation{}, err
	}
	return populationByWorldId(population, worldId)
}
