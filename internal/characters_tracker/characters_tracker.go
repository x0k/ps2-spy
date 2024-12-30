package characters_tracker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Tracker struct {
	wg        sync.WaitGroup
	platforms map[ps2_platforms.Platform]*platformTracker
}

func New(
	log *logger.Logger,
	charactersLoader CharacterLoader,
	publisher pubsub.Publisher[Event],
	mt *metrics.Metrics,
) *Tracker {
	platforms := make(map[ps2_platforms.Platform]*platformTracker, len(ps2_platforms.Platforms))
	for _, platform := range ps2_platforms.Platforms {
		platforms[platform] = newCharactersTracker(
			log.With(slog.String("platform", string(platform))),
			platform,
			ps2.PlatformWorldIds[platform],
			charactersLoader,
			publisher,
			mt,
		)
	}
	return &Tracker{
		platforms: platforms,
	}
}

func (t *Tracker) Start(ctx context.Context) {
	t.wg.Add(len(t.platforms))
	for _, tracker := range t.platforms {
		go func() {
			defer t.wg.Done()
			tracker.Start(ctx)
		}()
	}
	<-ctx.Done()
	t.wg.Wait()
}

func (t *Tracker) HandleLogin(
	ctx context.Context, platform ps2_platforms.Platform, event events.PlayerLogin,
) {
	t.platforms[platform].HandleLogin(ctx, event)
}

func (t *Tracker) HandleLogout(
	ctx context.Context, platform ps2_platforms.Platform, event events.PlayerLogout,
) {
	t.platforms[platform].HandleLogout(ctx, event)
}

func (t *Tracker) HandleWorldZoneAction(
	ctx context.Context, platform ps2_platforms.Platform, worldId, zoneId, charId string,
) {
	t.platforms[platform].HandleWorldZoneAction(ctx, worldId, zoneId, charId)
}

func (t *Tracker) OnlineOutfitMembers(
	ctx context.Context, platform ps2_platforms.Platform, outfitIds []ps2.OutfitId,
) (map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character, error) {
	return t.platforms[platform].OutfitMembersOnline(outfitIds), nil
}

func (t *Tracker) OnlineCharacters(
	ctx context.Context, platform ps2_platforms.Platform, characterIds []ps2.CharacterId,
) (map[ps2.CharacterId]ps2.Character, error) {
	return t.platforms[platform].CharactersOnline(characterIds), nil
}

func (t *Tracker) WorldsPopulation(platform ps2_platforms.Platform) ps2.WorldsPopulation {
	return t.platforms[platform].WorldsPopulation()
}

func (t *Tracker) DetailedWorldPopulation(platform ps2_platforms.Platform, worldId ps2.WorldId) (ps2.DetailedWorldPopulation, error) {
	return t.platforms[platform].DetailedWorldPopulation(worldId)
}
