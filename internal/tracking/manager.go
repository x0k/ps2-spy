package tracking

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var ErrUnknownEvent = fmt.Errorf("unknown event")

type Manager struct {
	log                             *logger.Logger
	wg                              sync.WaitGroup
	charactersFilterMu              sync.RWMutex
	charactersFilter                map[ps2.CharacterId]int
	outfitsFilterMu                 sync.RWMutex
	outfitsFilter                   map[ps2.OutfitId]int
	characterLoader                 loader.Keyed[ps2.CharacterId, ps2.Character]
	characterTrackingChannelsLoader loader.Keyed[ps2.Character, []discord.Channel]
	trackableCharactersLoader       loader.Simple[[]ps2.CharacterId]
	outfitMembersLoader             loader.Keyed[ps2.OutfitId, []ps2.CharacterId]
	outfitTrackingChannelsLoader    loader.Keyed[ps2.OutfitId, []discord.Channel]
	trackableOutfitsLoader          loader.Simple[[]ps2.OutfitId]
	rebuildFiltersInterval          time.Duration
}

func New(
	log *logger.Logger,
	charLoader loader.Keyed[ps2.CharacterId, ps2.Character],
	characterTrackingChannelsLoader loader.Keyed[ps2.Character, []discord.Channel],
	trackableCharactersLoader loader.Simple[[]ps2.CharacterId],
	outfitMembersLoader loader.Keyed[ps2.OutfitId, []ps2.CharacterId],
	outfitTrackingChannelsLoader loader.Keyed[ps2.OutfitId, []discord.Channel],
	trackableOutfitsLoader loader.Simple[[]ps2.OutfitId],
) *Manager {
	return &Manager{
		log:                             log,
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

func (tm *Manager) Start(ctx context.Context) {
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

func (tm *Manager) ChannelIdsForCharacter(ctx context.Context, characterId ps2.CharacterId) ([]discord.Channel, error) {
	const op = "tracking_manager.TrackingManager.channelIdsForCharacter"
	trackersCount := tm.characterTrackersCount(characterId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			tm.log.Warn(ctx, "invalid character trackers count", slog.String("char_id", string(characterId)))
		}
		return nil, nil
	}
	char, err := tm.characterLoader(ctx, characterId)
	if err != nil {
		return nil, fmt.Errorf("%s load character %s: %w", op, characterId, err)
	}
	return tm.characterTrackingChannelsLoader(ctx, char)
}

func (tm *Manager) ChannelIdsForOutfit(ctx context.Context, outfitId ps2.OutfitId) ([]discord.Channel, error) {
	trackersCount := tm.outfitTrackersCount(outfitId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			tm.log.Warn(ctx, "invalid outfit trackers count", slog.String("outfit_id", string(outfitId)))
		}
		return nil, nil
	}
	return tm.outfitTrackingChannelsLoader(ctx, outfitId)
}

func (tm *Manager) TrackOutfitMember(charId ps2.CharacterId, outfitId ps2.OutfitId) {
	count := tm.outfitTrackersCount(outfitId)
	tm.considerCharacter(charId, count)
}

func (tm *Manager) UntrackOutfitMember(charId ps2.CharacterId, outfitId ps2.OutfitId) {
	count := tm.outfitTrackersCount(outfitId)
	tm.considerCharacter(charId, -count)
}

func (tm *Manager) HandleTrackingSettingsUpdate(ctx context.Context, update TrackingSettingsUpdated) {
	tm.wg.Add(1)
	go tm.handleTrackingSettingsUpdateTask(ctx, update)
}

func (m *Manager) handleTrackingSettingsUpdateTask(ctx context.Context, upd TrackingSettingsUpdated) {
	defer m.wg.Done()
	for _, charId := range upd.Diff.Characters.ToAdd {
		m.considerCharacter(charId, 1)
	}
	for _, charId := range upd.Diff.Characters.ToDel {
		m.considerCharacter(charId, -1)
	}
	for _, outfitId := range upd.Diff.Outfits.ToAdd {
		m.considerOutfit(outfitId, 1)
		members, err := m.outfitMembersLoader(ctx, outfitId)
		if err != nil {
			m.log.Error(ctx, "failed to load outfit members", sl.Err(err))
			continue
		}
		m.considerOutfitMembers(members, 1)
	}
	for _, outfitId := range upd.Diff.Outfits.ToDel {
		m.considerOutfit(outfitId, -1)
		members, err := m.outfitMembersLoader(ctx, outfitId)
		if err != nil {
			m.log.Error(ctx, "failed to load outfit members", sl.Err(err))
			continue
		}
		m.considerOutfitMembers(members, -1)
	}
}

func (tm *Manager) considerCharacter(charId ps2.CharacterId, delta int) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId] += delta
}

func (tm *Manager) considerOutfit(outfitId ps2.OutfitId, delta int) {
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	tm.outfitsFilter[outfitId] += delta
}

func (tm *Manager) considerOutfitMembers(members []ps2.CharacterId, delta int) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	for _, member := range members {
		tm.charactersFilter[member] += delta
	}
}

func (tm *Manager) outfitTrackersCount(outfitId ps2.OutfitId) int {
	tm.outfitsFilterMu.RLock()
	defer tm.outfitsFilterMu.RUnlock()
	return tm.outfitsFilter[outfitId]
}

func (tm *Manager) characterTrackersCount(charId ps2.CharacterId) int {
	tm.charactersFilterMu.RLock()
	defer tm.charactersFilterMu.RUnlock()
	return tm.charactersFilter[charId]
}

func (tm *Manager) rebuildCharactersFilterTask(ctx context.Context) {
	defer tm.wg.Done()
	chars, err := tm.trackableCharactersLoader(ctx)
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

func (tm *Manager) rebuildOutfitsFilterTask(ctx context.Context) {
	defer tm.wg.Done()
	outfits, err := tm.trackableOutfitsLoader(ctx)
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

func (tm *Manager) rebuildFilters(ctx context.Context) {
	tm.log.Debug(ctx, "rebuilding filters")
	tm.wg.Add(2)
	go tm.rebuildCharactersFilterTask(ctx)
	go tm.rebuildOutfitsFilterTask(ctx)
}
