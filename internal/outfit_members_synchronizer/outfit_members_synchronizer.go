package outfit_members_synchronizer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/shared"
)

type OutfitMembersSaver = func(ctx context.Context, outfit ps2.OutfitId, members []ps2.CharacterId) error

type OutfitMembersSynchronizer struct {
	log *logger.Logger
	wg  sync.WaitGroup
	// Loaders and saver are platform specific
	outfitMembersLoader    loader.Keyed[ps2.OutfitId, []ps2.CharacterId]
	outfitSyncAtLoader     loader.Keyed[ps2.OutfitId, time.Time]
	trackableOutfitsLoader loader.Simple[[]ps2.OutfitId]
	membersSaver           OutfitMembersSaver
	refreshInterval        time.Duration
	ticker                 *time.Ticker
}

func New(
	log *logger.Logger,
	trackableOutfitsLoader loader.Simple[[]ps2.OutfitId],
	outfitSyncAtLoader loader.Keyed[ps2.OutfitId, time.Time],
	outfitMembersLoader loader.Keyed[ps2.OutfitId, []ps2.CharacterId],
	membersSaver OutfitMembersSaver,
	refreshInterval time.Duration,
) *OutfitMembersSynchronizer {
	return &OutfitMembersSynchronizer{
		log:                    log,
		trackableOutfitsLoader: trackableOutfitsLoader,
		outfitSyncAtLoader:     outfitSyncAtLoader,
		outfitMembersLoader:    outfitMembersLoader,
		membersSaver:           membersSaver,
		refreshInterval:        refreshInterval,
	}
}

func (s *OutfitMembersSynchronizer) Start(ctx context.Context) {
	s.ticker = time.NewTicker(s.refreshInterval)
	defer s.ticker.Stop()
	s.syncOutfits(ctx)
	for {
		select {
		case <-ctx.Done():
			s.wg.Wait()
			return
		case <-s.ticker.C:
			s.syncOutfits(ctx)
		}
	}
}

func (s *OutfitMembersSynchronizer) SyncOutfit(ctx context.Context, outfitId ps2.OutfitId) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.syncOutfit(ctx, outfitId)
	}()
}

func (s *OutfitMembersSynchronizer) syncOutfits(ctx context.Context) {
	outfits, err := s.trackableOutfitsLoader(ctx)
	s.log.Info(ctx, "synchronizing", slog.Int("outfits", len(outfits)))
	if err != nil {
		s.log.Error(ctx, "failed to load trackable outfits", sl.Err(err))
		return
	}
	for _, outfit := range outfits {
		select {
		case <-ctx.Done():
			return
		default:
			s.syncOutfit(ctx, outfit)
		}
	}
}

func (s *OutfitMembersSynchronizer) syncOutfit(ctx context.Context, outfitId ps2.OutfitId) {
	log := s.log.With(slog.String("outfit_id", string(outfitId)))
	err := retryable.New(func(ctx context.Context) error {
		syncAt, err := s.outfitSyncAtLoader(ctx, outfitId)
		isNotFound := errors.Is(err, shared.ErrNotFound)
		if err != nil && !isNotFound {
			return fmt.Errorf("failed to load last sync time: %w", err)
		}
		if !isNotFound && time.Since(syncAt) < s.refreshInterval {
			log.Debug(ctx, "skipping sync")
			return nil
		}
		members, err := s.outfitMembersLoader(ctx, outfitId)
		log.Debug(ctx, "synchronizing", slog.Int("members", len(members)))
		if err != nil {
			return fmt.Errorf("failed to load members from census: %w", err)
		}
		if err := s.membersSaver(ctx, outfitId, members); err != nil {
			return fmt.Errorf("failed to save members: %w", err)
		}
		return nil
	}).Run(
		ctx,
		while.ErrorIsHere,
		while.RetryCountIsLessThan(3),
		perform.Log(s.log.Logger, slog.LevelDebug, "members sync failed, retrying"),
	)
	if err != nil {
		log.Error(ctx, "failed to sync", sl.Err(err))
	}
}
