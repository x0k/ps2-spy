package ps2

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type Service struct {
	worldsFallbackLoader *fallbackLoader[WorldsPopulation]
	worldsPopulation     *containers.QueriedLoadableValue[string, string, Loaded[WorldsPopulation]]
	worldsLoaders        []string
	worldFallbackLoader  *keyedFallbackLoader[WorldId, DetailedWorldPopulation]
	worldPopulation      *containers.QueriedLoadableValue[multiLoaderQuery[WorldId], string, Loaded[DetailedWorldPopulation]]
	worldLoaders         []string
	alertsFallbackLoader *fallbackLoader[Alerts]
	alerts               *containers.QueriedLoadableValue[string, string, Loaded[Alerts]]
	alertsLoaders        []string
}

func NewService(
	populationLoaders map[string]Loader[WorldsPopulation],
	populationLoadersPriority []string,
	worldPopulationLoaders map[string]KeyedLoader[WorldId, DetailedWorldPopulation],
	worldPopulationLoadersPriority []string,
	alertsLoaders map[string]Loader[Alerts],
	alertsLoadersPriority []string,
) *Service {
	worldsLoadersWithDefault := maps.Clone(populationLoaders)
	worldsFallbackLoader := NewFallbackLoader(
		"Worlds",
		populationLoaders,
		populationLoadersPriority,
	)
	worldsLoadersWithDefault["default"] = worldsFallbackLoader
	worldsMultiLoader := NewMultiLoader(worldsLoadersWithDefault)
	worldLoadersWithDefault := maps.Clone(worldPopulationLoaders)
	worldFallbackLoader := NewKeyedFallbackLoader(
		"World",
		worldPopulationLoaders,
		worldPopulationLoadersPriority,
	)
	worldLoadersWithDefault["default"] = worldFallbackLoader
	worldKeyedMultiLoader := NewKeyedMultiLoader(worldLoadersWithDefault)
	alertsLoadersWithDefault := maps.Clone(alertsLoaders)
	alertsFallbackLoader := NewFallbackLoader(
		"Alerts",
		alertsLoaders,
		alertsLoadersPriority,
	)
	alertsLoadersWithDefault["default"] = alertsFallbackLoader
	alertsMultiLoader := NewMultiLoader(alertsLoadersWithDefault)
	return &Service{
		worldsFallbackLoader: worldsFallbackLoader,
		worldsLoaders:        populationLoadersPriority,
		worldsPopulation:     containers.NewKeyedLoadableValue(worldsMultiLoader, len(populationLoaders)+1, time.Minute),
		worldFallbackLoader:  worldFallbackLoader,
		worldLoaders:         worldPopulationLoadersPriority,
		worldPopulation: containers.NewQueriedLoadableValue[multiLoaderQuery[WorldId], string, Loaded[DetailedWorldPopulation]](
			worldKeyedMultiLoader,
			len(worldPopulationLoaders)+1,
			time.Minute,
			func(q multiLoaderQuery[WorldId]) string { return fmt.Sprintf("%s:%d", q.loader, int(q.key)) },
		),
		alertsFallbackLoader: alertsFallbackLoader,
		alertsLoaders:        alertsLoadersPriority,
		alerts:               containers.NewKeyedLoadableValue(alertsMultiLoader, len(alertsLoaders)+1, time.Minute),
	}
}

func providerName(provider string) string {
	if provider == "" {
		return "default"
	}
	return provider
}

func (s *Service) Start() {
	s.worldsFallbackLoader.Start()
	s.worldFallbackLoader.Start()
	s.alertsFallbackLoader.Start()
}

func (s *Service) Stop() {
	s.worldsFallbackLoader.Stop()
	s.worldFallbackLoader.Stop()
	s.alertsFallbackLoader.Stop()
}

func (s *Service) PopulationLoaders() []string {
	return s.worldsLoaders
}

func (s *Service) Population(ctx context.Context, provider string) (Loaded[WorldsPopulation], error) {
	return s.worldsPopulation.Load(ctx, providerName(provider))
}

func (s *Service) PopulationByWorldIdProviders() []string {
	return s.worldLoaders
}

func (s *Service) PopulationByWorldId(ctx context.Context, worldId WorldId, provider string) (Loaded[DetailedWorldPopulation], error) {
	return s.worldPopulation.Load(ctx, multiLoaderQuery[WorldId]{
		loader: providerName(provider),
		key:    worldId,
	})
}

func (s *Service) AlertsLoaders() []string {
	return s.alertsLoaders
}

func (s *Service) Alerts(ctx context.Context, provider string) (Loaded[Alerts], error) {
	return s.alerts.Load(ctx, providerName(provider))
}

func (s *Service) AlertsByWorldId(ctx context.Context, provider string, worldId WorldId) (Loaded[Alerts], error) {
	loaded, err := s.alerts.Load(ctx, providerName(provider))
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
