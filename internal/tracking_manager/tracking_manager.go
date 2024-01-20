package tracking_manager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
)

var ErrUnknownEvent = fmt.Errorf("unknown event")

type TrackingManager struct {
	charactersFilterMu              sync.RWMutex
	charactersFilter                map[string]int
	outfitsFilterMu                 sync.RWMutex
	outfitsFilter                   map[string]int
	characterLoader                 loaders.KeyedLoader[string, ps2.Character]
	characterTrackingChannelsLoader loaders.KeyedLoader[ps2.Character, []string]
	trackableCharactersLoader       loaders.Loader[[]string]
	outfitTrackingChannelsLoader    loaders.KeyedLoader[string, []string]
	trackableOutfitsLoader          loaders.Loader[[]string]
	rebuildFiltersInterval          time.Duration
}

func New(
	charLoader loaders.KeyedLoader[string, ps2.Character],
	characterTrackingChannelsLoader loaders.KeyedLoader[ps2.Character, []string],
	trackableCharactersLoader loaders.Loader[[]string],
	outfitTrackingChannelsLoader loaders.KeyedLoader[string, []string],
	trackableOutfitsLoader loaders.Loader[[]string],
) *TrackingManager {
	return &TrackingManager{
		charactersFilter:                make(map[string]int),
		outfitsFilter:                   make(map[string]int),
		characterLoader:                 charLoader,
		characterTrackingChannelsLoader: characterTrackingChannelsLoader,
		trackableCharactersLoader:       trackableCharactersLoader,
		outfitTrackingChannelsLoader:    outfitTrackingChannelsLoader,
		trackableOutfitsLoader:          trackableOutfitsLoader,
		rebuildFiltersInterval:          time.Hour * 12,
	}
}

func (tm *TrackingManager) rebuildCharactersFilter(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.rebuildCharactersFilter"
	log := infra.OpLogger(ctx, op)
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
}

func (tm *TrackingManager) rebuildOutfitsFilter(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.rebuildOutfitsFilter"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	outfits, err := tm.trackableOutfitsLoader.Load(ctx)
	if err != nil {
		log.Error("failed to load trackable outfits", sl.Err(err))
		return
	}
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	tm.outfitsFilter = make(map[string]int, len(outfits))
	for _, outfit := range outfits {
		tm.outfitsFilter[outfit]++
	}
}

func (tm *TrackingManager) rebuildFilters(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.rebuildFilters"
	infra.OpLogger(ctx, op).Debug("rebuilding filters")
	wg.Add(2)
	go tm.rebuildCharactersFilter(ctx, wg)
	go tm.rebuildOutfitsFilter(ctx, wg)
}

func (tm *TrackingManager) rebuildTicker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(tm.rebuildFiltersInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.rebuildFilters(ctx, wg)
		}
	}
}

func (tm *TrackingManager) Start(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.Start"
	tm.rebuildFilters(ctx, wg)
	wg.Add(1)
	go tm.rebuildTicker(ctx, wg)
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
	return tm.characterTrackingChannelsLoader.Load(ctx, char)
}

func (tm *TrackingManager) outfitTrackersCount(outfitTag string) int {
	tm.outfitsFilterMu.RLock()
	defer tm.outfitsFilterMu.RUnlock()
	return tm.outfitsFilter[outfitTag]
}

func (tm *TrackingManager) channelIdsForOutfit(ctx context.Context, outfitTag string) ([]string, error) {
	const op = "tracking_manager.TrackingManager.channelIdsForOutfit"
	log := infra.OpLogger(ctx, op)
	trackersCount := tm.outfitTrackersCount(outfitTag)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			log.Warn("invalid outfit trackers count", slog.String("outfit_tag", outfitTag))
		}
		return nil, nil
	}
	return tm.outfitTrackingChannelsLoader.Load(ctx, outfitTag)
}

func (tm *TrackingManager) ChannelIds(ctx context.Context, event any) ([]string, error) {
	const op = "TrackingManager.ChannelIds"
	switch e := event.(type) {
	case ps2events.PlayerLogin:
		return tm.channelIdsForCharacter(ctx, e.CharacterID)
	case ps2events.PlayerLogout:
		return tm.channelIdsForCharacter(ctx, e.CharacterID)
	case outfit_members_saver.OutfitMembersUpdate:
		return tm.channelIdsForOutfit(ctx, e.OutfitTag)
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

func (tm *TrackingManager) TrackOutfitMember(charId string, outfitTag string) {
	const op = "tracking_manager.TrackingManager.TrackOutfitMember"
	count := tm.outfitTrackersCount(outfitTag)
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId] += count
}

func (tm *TrackingManager) UntrackOutfitMember(charId string, outfitTag string) {
	const op = "tracking_manager.TrackingManager.UntrackOutfitMember"
	count := tm.outfitTrackersCount(outfitTag)
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId] -= count
}

func (tm *TrackingManager) TrackOutfit(outfitTag string) {
	const op = "tracking_manager.TrackingManager.TrackOutfit"
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	tm.outfitsFilter[outfitTag]++
}

func (tm *TrackingManager) UntrackOutfit(outfitTag string) {
	const op = "tracking_manager.TrackingManager.UntrackOutfit"
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	tm.outfitsFilter[outfitTag]--
}
