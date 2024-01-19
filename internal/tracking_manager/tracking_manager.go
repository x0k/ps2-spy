package tracking_manager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var ErrUnknownEvent = fmt.Errorf("unknown event")

type TrackingManager struct {
	charactersFilterMu        sync.RWMutex
	charactersFilter          map[string]int
	characterLoader           loaders.KeyedLoader[string, ps2.Character]
	trackingChannelsLoader    loaders.KeyedLoader[ps2.Character, []string]
	trackableCharactersLoader loaders.Loader[[]string]
	outfitTrackersCount       loaders.KeyedLoader[string, int]
}

func New(
	charLoader loaders.KeyedLoader[string, ps2.Character],
	trackingChannelsLoader loaders.KeyedLoader[ps2.Character, []string],
	trackableCharactersLoader loaders.Loader[[]string],
	outfitTrackersCount loaders.KeyedLoader[string, int],
) *TrackingManager {
	return &TrackingManager{
		charactersFilter:          make(map[string]int),
		characterLoader:           charLoader,
		trackingChannelsLoader:    trackingChannelsLoader,
		trackableCharactersLoader: trackableCharactersLoader,
		outfitTrackersCount:       outfitTrackersCount,
	}
}

func (tm *TrackingManager) Start(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.Start"
	log := infra.OpLogger(ctx, op)
	wg.Add(1)
	go func() {
		defer wg.Done()
		chars, err := tm.trackableCharactersLoader.Load(ctx)
		if err != nil {
			log.Error("failed to load trackable characters", sl.Err(err))
			return
		}
		tm.charactersFilterMu.Lock()
		defer tm.charactersFilterMu.Unlock()
		tm.charactersFilter = make(map[string]int, len(chars))
		for _, char := range chars {
			tm.charactersFilter[char]++
		}
	}()
}

func (tm *TrackingManager) characterTrackersCount(charId string) int {
	tm.charactersFilterMu.RLock()
	defer tm.charactersFilterMu.RUnlock()
	return tm.charactersFilter[charId]
}

func (tm *TrackingManager) channelIdsForCharacter(ctx context.Context, characterId string) ([]string, error) {
	const op = "tracking_manager.TrackingManager.channelIdsForCharacter"
	log := infra.OpLogger(ctx, op)
	trackersCount := tm.characterTrackersCount(characterId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			log.Warn("invalid character trackers count", slog.String("char_id", characterId))
		}
		return nil, nil
	}
	char, err := tm.characterLoader.Load(ctx, characterId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return tm.trackingChannelsLoader.Load(ctx, char)
}

func (tm *TrackingManager) ChannelIds(ctx context.Context, event any) ([]string, error) {
	const op = "TrackingManager.ChannelIds"
	switch e := event.(type) {
	case ps2events.PlayerLogin:
		return tm.channelIdsForCharacter(ctx, e.CharacterID)
	case ps2events.PlayerLogout:
		return tm.channelIdsForCharacter(ctx, e.CharacterID)
	}
	return nil, fmt.Errorf("%s: %w", op, ErrUnknownEvent)
}

func (tm *TrackingManager) TrackCharacter(charId string) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId]++
}

func (tm *TrackingManager) UntrackCharacter(charId string) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId]--
}

func (tm *TrackingManager) TrackOutfitMember(ctx context.Context, charId string, outfitTag string) {
	const op = "tracking_manager.TrackingManager.TrackOutfitMember"
	log := infra.OpLogger(ctx, op)
	count, err := tm.outfitTrackersCount.Load(ctx, outfitTag)
	if err != nil {
		log.Error("failed to load trackers count for outfit", slog.String("outfit", outfitTag), sl.Err(err))
		return
	}
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId] += count
}

func (tm *TrackingManager) UntrackOutfitMember(ctx context.Context, charId string, outfitTag string) {
	const op = "tracking_manager.TrackingManager.UntrackOutfitMember"
	log := infra.OpLogger(ctx, op)
	count, err := tm.outfitTrackersCount.Load(ctx, outfitTag)
	if err != nil {
		log.Error("failed to load trackers count for outfit", slog.String("outfit", outfitTag), sl.Err(err))
		return
	}
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId] -= count
}
