package outfit_members_synchronizer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type Saver interface {
	Save(ctx context.Context, outfit ps2.OutfitId, members []ps2.CharacterId) error
}

type OutfitMembersSynchronizer struct {
	// Loaders and saver are platform specific
	censusMembersLoader    loaders.KeyedLoader[ps2.OutfitId, []ps2.CharacterId]
	outfitSyncAtLoader     loaders.KeyedLoader[ps2.OutfitId, time.Time]
	trackableOutfitsLoader loaders.Loader[[]ps2.OutfitId]
	membersSaver           Saver
	refreshInterval        time.Duration
	ticker                 *time.Ticker
	started                atomic.Bool
}

func New(
	trackableOutfitsLoader loaders.Loader[[]ps2.OutfitId],
	outfitSyncAtLoader loaders.KeyedLoader[ps2.OutfitId, time.Time],
	censusMembersLoader loaders.KeyedLoader[ps2.OutfitId, []ps2.CharacterId],
	membersSaver Saver,
	refreshInterval time.Duration,
) *OutfitMembersSynchronizer {
	return &OutfitMembersSynchronizer{
		trackableOutfitsLoader: trackableOutfitsLoader,
		outfitSyncAtLoader:     outfitSyncAtLoader,
		censusMembersLoader:    censusMembersLoader,
		membersSaver:           membersSaver,
		refreshInterval:        refreshInterval,
	}
}

func (s *OutfitMembersSynchronizer) saveMembersTask(ctx context.Context, wg *sync.WaitGroup, outfitId ps2.OutfitId, members []ps2.CharacterId) {
	const op = "outfit_members_synchronizer.OutfitMembersSynchronizer.saveMembersTask"
	log := infra.OpLogger(ctx, op).With(slog.String("outfit_id", string(outfitId)), slog.Int("members_count", len(members)))
	defer wg.Done()
	if err := s.membersSaver.Save(ctx, outfitId, members); err != nil {
		log.Error("failed to save members", sl.Err(err))
	}
}

func (s *OutfitMembersSynchronizer) SyncOutfit(ctx context.Context, wg *sync.WaitGroup, outfitId ps2.OutfitId) {
	const op = "outfit_members_synchronizer.OutfitMembersSynchronizer.SyncOutfit"
	log := infra.Logger(ctx).With(infra.Op(op), slog.String("outfit_id", string(outfitId)))
	err := retryable.New(func(ctx context.Context) error {
		syncAt, err := s.outfitSyncAtLoader.Load(ctx, outfitId)
		isNotFound := errors.Is(err, loaders.ErrNotFound)
		if err != nil && !isNotFound {
			return fmt.Errorf("failed to load last sync time: %w", err)
		}
		if !isNotFound && time.Since(syncAt) < s.refreshInterval {
			log.Debug("skipping sync")
			return nil
		}
		members, err := s.censusMembersLoader.Load(ctx, outfitId)
		log.Debug("synchronizing", slog.Int("members", len(members)))
		if err != nil {
			return fmt.Errorf("failed to load members from census: %w", err)
		}
		wg.Add(1)
		go s.saveMembersTask(ctx, wg, outfitId, members)
		return nil
	}).Run(
		ctx,
		while.ErrorIsHere,
		while.RetryCountIsLessThan(3),
		perform.Debug(log, "retry to load members"),
	)
	if err != nil {
		log.Error("failed to sync", sl.Err(err))
	}
}

func (s *OutfitMembersSynchronizer) sync(ctx context.Context, wg *sync.WaitGroup) {
	const op = "outfit_members_synchronizer.OutfitMembersSynchronizer.sync"
	log := infra.OpLogger(ctx, op)
	outfits, err := s.trackableOutfitsLoader.Load(ctx)
	log.Info("synchronizing", slog.Int("outfits", len(outfits)))
	if err != nil {
		log.Error("failed to load trackable outfits", sl.Err(err))
		return
	}
	for _, outfit := range outfits {
		select {
		case <-ctx.Done():
			return
		default:
			s.SyncOutfit(ctx, wg, outfit)
		}
	}
}

func (s *OutfitMembersSynchronizer) Start(ctx context.Context, wg *sync.WaitGroup) {
	const op = "outfit_members_synchronizer.OutfitMembersSynchronizer.Start"
	if s.started.Swap(true) {
		return
	}
	infra.OpLogger(ctx, op).Info("starting")
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.sync(ctx, wg)
		s.ticker = time.NewTicker(s.refreshInterval)
		defer s.ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.ticker.C:
				s.sync(ctx, wg)
			}
		}
	}()
}
