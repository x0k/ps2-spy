package tracking_manager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/facilities_manager"
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
	charactersFilter                map[ps2.CharacterId]int
	outfitsFilterMu                 sync.RWMutex
	outfitsFilter                   map[ps2.OutfitId]int
	characterLoader                 loaders.KeyedLoader[ps2.CharacterId, ps2.Character]
	characterTrackingChannelsLoader loaders.KeyedLoader[ps2.Character, []string]
	trackableCharactersLoader       loaders.Loader[[]ps2.CharacterId]
	outfitMembersLoader             loaders.KeyedLoader[ps2.OutfitId, []ps2.CharacterId]
	outfitTrackingChannelsLoader    loaders.KeyedLoader[ps2.OutfitId, []string]
	trackableOutfitsLoader          loaders.Loader[[]ps2.OutfitId]
	rebuildFiltersInterval          time.Duration
}

func New(
	charLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	characterTrackingChannelsLoader loaders.KeyedLoader[ps2.Character, []string],
	trackableCharactersLoader loaders.Loader[[]ps2.CharacterId],
	outfitMembersLoader loaders.KeyedLoader[ps2.OutfitId, []ps2.CharacterId],
	outfitTrackingChannelsLoader loaders.KeyedLoader[ps2.OutfitId, []string],
	trackableOutfitsLoader loaders.Loader[[]ps2.OutfitId],
) *TrackingManager {
	return &TrackingManager{
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

func (tm *TrackingManager) rebuildCharactersFilterTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.rebuildCharactersFilterTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	chars, err := tm.trackableCharactersLoader.Load(ctx)
	if err != nil {
		log.Error("failed to load trackable characters", sl.Err(err))
		return
	}
	tm.charactersFilterMu.Lock()
	defer tm.charactersFilterMu.Unlock()
	oldSize := len(tm.charactersFilter)
	newSize := len(chars)
	if oldSize != newSize {
		log.Info(
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
	const op = "tracking_manager.TrackingManager.rebuildOutfitsFilterTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	outfits, err := tm.trackableOutfitsLoader.Load(ctx)
	if err != nil {
		log.Error("failed to load trackable outfits", sl.Err(err))
		return
	}
	tm.outfitsFilterMu.Lock()
	defer tm.outfitsFilterMu.Unlock()
	oldSize := len(tm.outfitsFilter)
	newSize := len(outfits)
	if oldSize != newSize {
		log.Info(
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

func (tm *TrackingManager) goRebuildFiltersTasks(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.goRebuildFiltersTasks"
	infra.OpLogger(ctx, op).Debug("rebuilding filters")
	wg.Add(2)
	go tm.rebuildCharactersFilterTask(ctx, wg)
	go tm.rebuildOutfitsFilterTask(ctx, wg)
}

func (tm *TrackingManager) scheduleFiltersRebuildTasks(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(tm.rebuildFiltersInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.goRebuildFiltersTasks(ctx, wg)
		}
	}
}

func (tm *TrackingManager) Start(ctx context.Context, wg *sync.WaitGroup) {
	const op = "tracking_manager.TrackingManager.Start"
	tm.goRebuildFiltersTasks(ctx, wg)
	wg.Add(1)
	go tm.scheduleFiltersRebuildTasks(ctx, wg)
}

func (tm *TrackingManager) characterTrackersCount(charId ps2.CharacterId) int {
	tm.charactersFilterMu.RLock()
	defer tm.charactersFilterMu.RUnlock()
	return tm.charactersFilter[charId]
}

func (tm *TrackingManager) channelIdsForCharacter(ctx context.Context, characterId ps2.CharacterId) ([]string, error) {
	const op = "tracking_manager.TrackingManager.channelIdsForCharacter"
	log := infra.OpLogger(ctx, op)
	trackersCount := tm.characterTrackersCount(characterId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			log.Warn("invalid character trackers count", slog.String("charId", string(characterId)))
		}
		return nil, nil
	}
	char, err := tm.characterLoader.Load(ctx, characterId)
	if err != nil {
		return nil, fmt.Errorf("%s load character %s: %w", op, characterId, err)
	}
	return tm.characterTrackingChannelsLoader.Load(ctx, char)
}

func (tm *TrackingManager) outfitTrackersCount(outfitId ps2.OutfitId) int {
	tm.outfitsFilterMu.RLock()
	defer tm.outfitsFilterMu.RUnlock()
	return tm.outfitsFilter[outfitId]
}

func (tm *TrackingManager) channelIdsForOutfit(ctx context.Context, outfitId ps2.OutfitId) ([]string, error) {
	const op = "tracking_manager.TrackingManager.channelIdsForOutfitTag"
	log := infra.OpLogger(ctx, op)
	trackersCount := tm.outfitTrackersCount(outfitId)
	if trackersCount <= 0 {
		if trackersCount < 0 {
			log.Warn("invalid outfit trackers count", slog.String("outfitId", string(outfitId)))
		}
		return nil, nil
	}
	return tm.outfitTrackingChannelsLoader.Load(ctx, outfitId)
}

func (tm *TrackingManager) ChannelIds(ctx context.Context, event any) ([]string, error) {
	const op = "TrackingManager.ChannelIds"
	switch e := event.(type) {
	case ps2events.PlayerLogin:
		return tm.channelIdsForCharacter(ctx, ps2.CharacterId(e.CharacterID))
	case ps2events.PlayerLogout:
		return tm.channelIdsForCharacter(ctx, ps2.CharacterId(e.CharacterID))
	case outfit_members_saver.OutfitMembersUpdate:
		return tm.channelIdsForOutfit(ctx, e.OutfitId)
	case facilities_manager.FacilityControl:
		return tm.channelIdsForOutfit(ctx, ps2.OutfitId(e.OutfitID))
	case facilities_manager.FacilityLoss:
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
	const op = "tracking_manager.TrackingManager.TrackOutfitMember"
	count := tm.outfitTrackersCount(outfitId)
	tm.considerCharacter(charId, count)
}

func (tm *TrackingManager) UntrackOutfitMember(charId ps2.CharacterId, outfitId ps2.OutfitId) {
	const op = "tracking_manager.TrackingManager.UntrackOutfitMember"
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
	members, err := tm.outfitMembersLoader.Load(ctx, outfitId)
	if err != nil {
		return fmt.Errorf("%s load members of %q: %w", op, outfitId, err)
	}
	tm.considerOutfitMembers(members, 1)
	return nil
}

func (tm *TrackingManager) UntrackOutfit(ctx context.Context, outfitId ps2.OutfitId) error {
	const op = "tracking_manager.TrackingManager.UntrackOutfit"
	tm.considerOutfit(outfitId, -1)
	members, err := tm.outfitMembersLoader.Load(ctx, outfitId)
	if err != nil {
		return fmt.Errorf("%s load members of %q: %w", op, outfitId, err)
	}
	tm.considerOutfitMembers(members, -1)
	return nil
}
