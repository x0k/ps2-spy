package tracking_manager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	sql_outfit_members_saver "github.com/x0k/ps2-spy/internal/savers/outfit_members/sql"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

var ErrUnknownEvent = fmt.Errorf("unknown event")

type TrackingManager struct {
	name                            string
	log                             *logger.Logger
	charactersFilterMu              sync.RWMutex
	charactersFilter                map[ps2.CharacterId]int
	outfitsFilterMu                 sync.RWMutex
	outfitsFilter                   map[ps2.OutfitId]int
	characterLoader                 loader.Keyed[ps2.CharacterId, ps2.Character]
	characterTrackingChannelsLoader loader.Keyed[ps2.Character, []meta.ChannelId]
	trackableCharactersLoader       loader.Simple[[]ps2.CharacterId]
	outfitMembersLoader             loader.Keyed[ps2.OutfitId, []ps2.CharacterId]
	outfitTrackingChannelsLoader    loader.Keyed[ps2.OutfitId, []meta.ChannelId]
	trackableOutfitsLoader          loader.Simple[[]ps2.OutfitId]
	rebuildFiltersInterval          time.Duration
}

func New(
	name string,
	log *logger.Logger,
	charLoader loader.Keyed[ps2.CharacterId, ps2.Character],
	characterTrackingChannelsLoader loader.Keyed[ps2.Character, []meta.ChannelId],
	trackableCharactersLoader loader.Simple[[]ps2.CharacterId],
	outfitMembersLoader loader.Keyed[ps2.OutfitId, []ps2.CharacterId],
	outfitTrackingChannelsLoader loader.Keyed[ps2.OutfitId, []meta.ChannelId],
	trackableOutfitsLoader loader.Simple[[]ps2.OutfitId],
) *TrackingManager {
	return &TrackingManager{
		name:                            name,
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

func (tm *TrackingManager) Name() string {
	return tm.name
}

func (tm *TrackingManager) rebuildCharactersFilterTask(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
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

func (tm *TrackingManager) rebuildOutfitsFilterTask(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
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

func (tm *TrackingManager) rebuildFilters(ctx context.Context, wg *sync.WaitGroup) {
	tm.log.Debug(ctx, "rebuilding filters")
	wg.Add(2)
	go tm.rebuildCharactersFilterTask(ctx, wg)
	go tm.rebuildOutfitsFilterTask(ctx, wg)
}

func (tm *TrackingManager) Start(ctx context.Context) error {
	ticker := time.NewTicker(tm.rebuildFiltersInterval)
	defer ticker.Stop()
	wg := &sync.WaitGroup{}
	tm.rebuildFilters(ctx, wg)
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		case <-ticker.C:
			tm.rebuildFilters(ctx, wg)
		}
	}
}

func (tm *TrackingManager) characterTrackersCount(charId ps2.CharacterId) int {
	tm.charactersFilterMu.RLock()
	defer tm.charactersFilterMu.RUnlock()
	return tm.charactersFilter[charId]
}

func (tm *TrackingManager) channelIdsForCharacter(ctx context.Context, characterId ps2.CharacterId) ([]meta.ChannelId, error) {
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

func (tm *TrackingManager) outfitTrackersCount(outfitId ps2.OutfitId) int {
	tm.outfitsFilterMu.RLock()
	defer tm.outfitsFilterMu.RUnlock()
	return tm.outfitsFilter[outfitId]
}

func (tm *TrackingManager) channelIdsForOutfit(ctx context.Context, outfitId ps2.OutfitId) ([]meta.ChannelId, error) {
	trackersCount := tm.outfitTrackersCount(outfitId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			tm.log.Warn(ctx, "invalid outfit trackers count", slog.String("outfit_id", string(outfitId)))
		}
		return nil, nil
	}
	return tm.outfitTrackingChannelsLoader(ctx, outfitId)
}

func (tm *TrackingManager) ChannelIds(ctx context.Context, event any) ([]meta.ChannelId, error) {
	const op = "TrackingManager.ChannelIds"
	switch e := event.(type) {
	case characters_tracker.PlayerLogin:
		return tm.channelIdsForCharacter(ctx, e.CharacterId)
	case characters_tracker.PlayerLogout:
		return tm.channelIdsForCharacter(ctx, e.CharacterId)
	case sql_outfit_members_saver.OutfitMembersUpdate:
		return tm.channelIdsForOutfit(ctx, e.OutfitId)
	case worlds_tracker.FacilityControl:
		return tm.channelIdsForOutfit(ctx, ps2.OutfitId(e.OutfitID))
	case worlds_tracker.FacilityLoss:
		return tm.channelIdsForOutfit(ctx, e.OldOutfitId)
	}
	return nil, fmt.Errorf("%s: %w", op, ErrUnknownEvent)
}

func (tm *TrackingManager) considerCharacter(charId ps2.CharacterId, delta int) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	tm.charactersFilter[charId] += delta
}

func (tm *TrackingManager) TrackCharacter(charId ps2.CharacterId) {
	tm.considerCharacter(charId, 1)
}

func (tm *TrackingManager) UntrackCharacter(charId ps2.CharacterId) {
	tm.considerCharacter(charId, -1)
}

func (tm *TrackingManager) TrackOutfitMember(charId ps2.CharacterId, outfitId ps2.OutfitId) {
	count := tm.outfitTrackersCount(outfitId)
	tm.considerCharacter(charId, count)
}

func (tm *TrackingManager) UntrackOutfitMember(charId ps2.CharacterId, outfitId ps2.OutfitId) {
	count := tm.outfitTrackersCount(outfitId)
	tm.considerCharacter(charId, -count)
}

func (tm *TrackingManager) considerOutfit(outfitId ps2.OutfitId, delta int) {
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	tm.outfitsFilter[outfitId] += delta
}

func (tm *TrackingManager) considerOutfitMembers(members []ps2.CharacterId, delta int) {
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	for _, member := range members {
		tm.charactersFilter[member] += delta
	}
}

func (tm *TrackingManager) TrackOutfit(ctx context.Context, outfitId ps2.OutfitId) error {
	const op = "tracking_manager.TrackingManager.TrackOutfit"
	tm.considerOutfit(outfitId, 1)
	members, err := tm.outfitMembersLoader(ctx, outfitId)
	if err != nil {
		return fmt.Errorf("%s load members of %q: %w", op, outfitId, err)
	}
	tm.considerOutfitMembers(members, 1)
	return nil
}

func (tm *TrackingManager) UntrackOutfit(ctx context.Context, outfitId ps2.OutfitId) error {
	const op = "tracking_manager.TrackingManager.UntrackOutfit"
	tm.considerOutfit(outfitId, -1)
	members, err := tm.outfitMembersLoader(ctx, outfitId)
	if err != nil {
		return fmt.Errorf("%s load members of %q: %w", op, outfitId, err)
	}
	tm.considerOutfitMembers(members, -1)
	return nil
}
