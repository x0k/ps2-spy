package outfit_members_synchronizer

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retry"
	"github.com/x0k/ps2-spy/internal/loaders"
)

type Saver interface {
	Save(ctx context.Context, outfit string, members []string) error
}

type OutfitMembersSynchronizer struct {
	log *slog.Logger
	// Loaders and saver are platform specific
	censusMembersLoader    loaders.KeyedLoader[string, []string]
	outfitSyncAtLoader     loaders.KeyedLoader[string, time.Time]
	trackableOutfitsLoader loaders.Loader[[]string]
	membersSaver           Saver
	refreshInterval        time.Duration
	ticker                 *time.Ticker
	started                atomic.Bool
}

func New(
	log *slog.Logger,
	trackableOutfitsLoader loaders.Loader[[]string],
	outfitSyncAtLoader loaders.KeyedLoader[string, time.Time],
	censusMembersLoader loaders.KeyedLoader[string, []string],
	membersSaver Saver,
	refreshInterval time.Duration,
) *OutfitMembersSynchronizer {
	return &OutfitMembersSynchronizer{
		log:                    log.With(slog.String("component", "outfit_members_synchronizer")),
		trackableOutfitsLoader: trackableOutfitsLoader,
		outfitSyncAtLoader:     outfitSyncAtLoader,
		censusMembersLoader:    censusMembersLoader,
		membersSaver:           membersSaver,
		refreshInterval:        refreshInterval,
	}
}

func (s *OutfitMembersSynchronizer) saveMembers(ctx context.Context, wg *sync.WaitGroup, outfitTag string, members []string) {
	defer wg.Done()
	if err := s.membersSaver.Save(ctx, outfitTag, members); err != nil {
		s.log.Error("failed to save members", slog.String("outfit", outfitTag), sl.Err(err))
	}
}

func (s *OutfitMembersSynchronizer) SyncOutfit(ctx context.Context, wg *sync.WaitGroup, outfitTag string) {
	retry.RetryWhileWithRecover(retry.Retryable{
		Try: func() error {
			syncAt, err := s.outfitSyncAtLoader.Load(ctx, outfitTag)
			isNotFound := errors.Is(err, loaders.ErrNotFound)
			if err != nil && !isNotFound {
				s.log.Error("failed to load last sync time", slog.String("outfit", outfitTag), sl.Err(err))
				return err
			}
			if !isNotFound && time.Since(syncAt) < s.refreshInterval {
				s.log.Debug("skipping sync", slog.String("outfit", outfitTag))
				return nil
			}
			members, err := s.censusMembersLoader.Load(ctx, outfitTag)
			s.log.Debug("synchronizing", slog.String("outfit", outfitTag), slog.Int("members", len(members)))
			if err != nil {
				s.log.Error("failed to load members from census", slog.String("outfit", outfitTag), sl.Err(err))
				return err
			}
			wg.Add(1)
			go s.saveMembers(ctx, wg, outfitTag, members)
			// s.saveMembers(ctx, wg, outfitTag, members)
			return nil
		},
		While: retry.ContextIsNotCanceledAndMaxRetriesNotExceeded(3),
		BeforeSleep: func(d time.Duration) {
			s.log.Debug("retry to load members", slog.String("outfit", outfitTag), slog.Duration("after", d))
		},
	})
}

func (s *OutfitMembersSynchronizer) sync(ctx context.Context, wg *sync.WaitGroup) {
	outfits, err := s.trackableOutfitsLoader.Load(ctx)
	s.log.Info("synchronizing", slog.Int("outfits", len(outfits)))
	if err != nil {
		s.log.Error("failed to load trackable outfits", sl.Err(err))
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
	if s.started.Swap(true) {
		return
	}
	s.log.Debug("started")
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
