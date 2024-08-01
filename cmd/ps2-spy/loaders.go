package main

import (
	"context"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/loaders/characters_loader"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func newCharactersLoader(
	log *logger.Logger,
	mt metrics.Metrics,
	platform platforms.Platform,
	censusClient *census2.Client,
) loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character] {
	return metrics.InstrumentMultiKeyedLoaderWithSubjectsCounter(
		mt.PlatformLoaderSubjectsCounterMetric(metrics.CharactersPlatformLoaderName, platform),
		characters_loader.NewCensus(log, censusClient, platform),
	)
}

func startNewBatchedCharacterLoader(
	ctx context.Context,
	wg *sync.WaitGroup,
	log *logger.Logger,
	mt metrics.Metrics,
	platform platforms.Platform,
	loader loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character],
) loaders.KeyedLoader[ps2.CharacterId, ps2.Character] {
	batched := loaders.NewBatchLoader(loader, 10*time.Second)
	batched.Start(ctx, wg)
	return loaders.NewCachedQueriedLoader(
		log.Logger,
		metrics.InstrumentQueriedLoaderWithCounterMetric(
			mt.PlatformLoadsCounterMetric(metrics.CharacterPlatformLoaderName, platform),
			batched,
		),
		containers.NewExpiableLRU[ps2.CharacterId, ps2.Character](0, 24*time.Hour),
	)
}
