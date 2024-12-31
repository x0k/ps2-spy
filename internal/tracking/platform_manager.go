package tracking

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrUnknownEvent = fmt.Errorf("unknown event")

type CharacterLoader = func(
	ctx context.Context, platform ps2_platforms.Platform, characterId ps2.CharacterId,
) (ps2.Character, error)
type CharacterTrackingChannelsLoader = func(
	ctx context.Context, platform ps2_platforms.Platform, character ps2.Character,
) ([]discord.Channel, error)
type TrackableCharactersLoader = func(
	context.Context, ps2_platforms.Platform,
) ([]ps2.CharacterId, error)
type OutfitMembersLoader = func(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId,
) ([]ps2.CharacterId, error)
type OutfitTrackingChannelsLoader = func(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId,
) ([]discord.Channel, error)
type TrackableOutfitsLoader = func(
	context.Context, ps2_platforms.Platform,
) ([]ps2.OutfitId, error)

type platformManager struct {
	log                             *logger.Logger
	platform                        ps2_platforms.Platform
	wg                              sync.WaitGroup
	charactersFilterMu              sync.RWMutex
	charactersFilter                map[ps2.CharacterId]int
	outfitsFilterMu                 sync.RWMutex
	outfitsFilter                   map[ps2.OutfitId]int
	characterLoader                 CharacterLoader
	characterTrackingChannelsLoader CharacterTrackingChannelsLoader
	trackableCharactersLoader       TrackableCharactersLoader
	outfitMembersLoader             OutfitMembersLoader
	outfitTrackingChannelsLoader    OutfitTrackingChannelsLoader
	trackableOutfitsLoader          TrackableOutfitsLoader
	rebuildFiltersInterval          time.Duration
}

func newPlatformManager(
	log *logger.Logger,
	platform ps2_platforms.Platform,
	charLoader CharacterLoader,
	characterTrackingChannelsLoader CharacterTrackingChannelsLoader,
	trackableCharactersLoader TrackableCharactersLoader,
	outfitMembersLoader OutfitMembersLoader,
	outfitTrackingChannelsLoader OutfitTrackingChannelsLoader,
	trackableOutfitsLoader TrackableOutfitsLoader,
) *platformManager {
	return &platformManager{
		log:                             log,
		platform:                        platform,
		charactersFilter:                make(map[ps2.CharacterId]int),
		outfitsFilter:                   make(map[ps2.OutfitId]int),
		characterLoader:                 charLoader,
		characterTrackingChannelsLoader: characterTrackingChannelsLoader,
		trackableCharactersLoader:       trackableCharactersLoader,
		outfitMembersLoader:             outfitMembersLoader,
		outfitTrackingChannelsLoader:    outfitTrackingChannelsLoader,
		trackableOutfitsLoader:          trackableOutfitsLoader,
		rebuildFiltersInterval:          time.Hour * 12,
	}
}

func (tm *platformManager) Start(ctx context.Context) {
	ticker := time.NewTicker(tm.rebuildFiltersInterval)
	defer ticker.Stop()
	tm.rebuildFilters(ctx)
	for {
		select {
		case <-ctx.Done():
			tm.wg.Wait()
			return
		case <-ticker.C:
			tm.rebuildFilters(ctx)
		}
	}
}

func (tm *platformManager) ChannelIdsForCharacter(ctx context.Context, characterId ps2.CharacterId) ([]discord.Channel, error) {
	const op = "tracking_manager.TrackingManager.channelIdsForCharacter"
	trackersCount := tm.characterTrackersCount(characterId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			tm.log.Warn(ctx, "invalid character trackers count", slog.String("char_id", string(characterId)))
		}
		return nil, nil
	}
	char, err := tm.characterLoader(ctx, tm.platform, characterId)
	if err != nil {
		return nil, fmt.Errorf("%s load character %s: %w", op, characterId, err)
	}
	return tm.characterTrackingChannelsLoader(ctx, tm.platform, char)
}

func (tm *platformManager) ChannelIdsForOutfit(ctx context.Context, outfitId ps2.OutfitId) ([]discord.Channel, error) {
	trackersCount := tm.outfitTrackersCount(outfitId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			tm.log.Warn(ctx, "invalid outfit trackers count", slog.String("outfit_id", string(outfitId)))
		}
		return nil, nil
	}
	return tm.outfitTrackingChannelsLoader(ctx, tm.platform, outfitId)
}

func (tm *platformManager) TrackOutfitMembers(outfitId ps2.OutfitId, charIds []ps2.CharacterId) {
	count := tm.outfitTrackersCount(outfitId)
	for _, charId := range charIds {
		tm.considerCharacter(charId, count)
	}
}

func (tm *platformManager) UntrackOutfitMembers(outfitId ps2.OutfitId, charIds []ps2.CharacterId) {
	count := tm.outfitTrackersCount(outfitId)
	for _, charId := range charIds {
		tm.considerCharacter(charId, -count)
	}
}

func (tm *platformManager) HandleTrackingSettingsUpdate(ctx context.Context, update TrackingSettingsUpdated) {
	tm.wg.Add(1)
	go tm.handleTrackingSettingsUpdateTask(ctx, update)
}

func (m *platformManager) handleTrackingSettingsUpdateTask(ctx context.Context, upd TrackingSettingsUpdated) {
	defer m.wg.Done()
	for _, charId := range upd.Diff.Characters.ToAdd {
		m.considerCharacter(charId, 1)
	}
	for _, charId := range upd.Diff.Characters.ToDel {
		m.considerCharacter(charId, -1)
	}
	for _, outfitId := range upd.Diff.Outfits.ToAdd {
		m.considerOutfit(outfitId, 1)
		members, err := m.outfitMembersLoader(ctx, m.platform, outfitId)
		if err != nil {
			m.log.Error(ctx, "failed to load outfit members", sl.Err(err))
			continue
		}
		m.considerOutfitMembers(members, 1)
	}
	for _, outfitId := range upd.Diff.Outfits.ToDel {
		m.considerOutfit(outfitId, -1)
		members, err := m.outfitMembersLoader(ctx, m.platform, outfitId)
		if err != nil {
			m.log.Error(ctx, "failed to load outfit members", sl.Err(err))
			continue
		}
		m.considerOutfitMembers(members, -1)
	}
}

func (tm *platformManager) considerCharacter(charId ps2.CharacterId, delta int) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId] += delta
}

func (tm *platformManager) considerOutfit(outfitId ps2.OutfitId, delta int) {
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	tm.outfitsFilter[outfitId] += delta
}

func (tm *platformManager) considerOutfitMembers(members []ps2.CharacterId, delta int) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	for _, member := range members {
		tm.charactersFilter[member] += delta
	}
}

func (tm *platformManager) outfitTrackersCount(outfitId ps2.OutfitId) int {
	tm.outfitsFilterMu.RLock()
	defer tm.outfitsFilterMu.RUnlock()
	return tm.outfitsFilter[outfitId]
}

func (tm *platformManager) characterTrackersCount(charId ps2.CharacterId) int {
	tm.charactersFilterMu.RLock()
	defer tm.charactersFilterMu.RUnlock()
	return tm.charactersFilter[charId]
}

func (tm *platformManager) rebuildCharactersFilterTask(ctx context.Context) {
	defer tm.wg.Done()
	chars, err := tm.trackableCharactersLoader(ctx, tm.platform)
	if err != nil {
		tm.log.Error(ctx, "failed to load trackable characters", sl.Err(err))
		return
	}
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	oldSize := len(tm.charactersFilter)
	newSize := len(chars)
	if oldSize != newSize {
		tm.log.Info(
			ctx,
			"fixing inconsistent characters filter",
			slog.Int("old_size", oldSize),
			slog.Int("new_size", newSize),
		)
	}
	tm.charactersFilter = make(map[ps2.CharacterId]int, len(chars))
	for _, char := range chars {
		tm.charactersFilter[char]++
	}
}

func (tm *platformManager) rebuildOutfitsFilterTask(ctx context.Context) {
	defer tm.wg.Done()
	outfits, err := tm.trackableOutfitsLoader(ctx, tm.platform)
	if err != nil {
		tm.log.Error(ctx, "failed to load trackable outfits", sl.Err(err))
		return
	}
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	oldSize := len(tm.outfitsFilter)
	newSize := len(outfits)
	if oldSize != newSize {
		tm.log.Info(
			ctx,
			"fixing inconsistent outfits filter",
			slog.Int("old_size", oldSize),
			slog.Int("new_size", newSize),
		)
	}
	tm.outfitsFilter = make(map[ps2.OutfitId]int, len(outfits))
	for _, outfit := range outfits {
		tm.outfitsFilter[outfit]++
	}
}

func (tm *platformManager) rebuildFilters(ctx context.Context) {
	tm.log.Debug(ctx, "rebuilding filters")
	tm.wg.Add(2)
	go tm.rebuildCharactersFilterTask(ctx)
	go tm.rebuildOutfitsFilterTask(ctx)
}
