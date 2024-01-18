package outfit_members_synchronizer

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
)

type Saver interface {
	Save(ctx context.Context, outfit string, members []string) error
}

type OutfitMembersSynchronizer struct {
	log *slog.Logger
	// Loaders and saver are platform specific
	censusLoader    loaders.KeyedLoader[string, []string]
	membersSaver    Saver
	outfitsLoader   loaders.Loader[[]string]
	refreshInterval time.Duration
	ticker          *time.Ticker
	started         atomic.Bool
}

func New(
	log *slog.Logger,
	trackableOutfitsLoader loaders.Loader[[]string],
	censusMembersLoader loaders.KeyedLoader[string, []string],
	membersSaver Saver,
	refreshInterval time.Duration,
) *OutfitMembersSynchronizer {
	return &OutfitMembersSynchronizer{
		log:             log.With(slog.String("component", "outfit_members_synchronizer")),
		censusLoader:    censusMembersLoader,
		outfitsLoader:   trackableOutfitsLoader,
		membersSaver:    membersSaver,
		refreshInterval: refreshInterval,
	}
}

func (s *OutfitMembersSynchronizer) saveMembers(ctx context.Context, wg *sync.WaitGroup, outfitTag string, members []string) {
	defer wg.Done()
	if err := s.membersSaver.Save(ctx, outfitTag, members); err != nil {
		s.log.Error("failed to save members", slog.String("outfit", outfitTag), sl.Err(err))
	}
}

func (s *OutfitMembersSynchronizer) SyncOutfit(ctx context.Context, wg *sync.WaitGroup, outfitTag string) {
	members, err := s.censusLoader.Load(ctx, outfitTag)
	s.log.Debug("synchronizing", slog.String("outfit", outfitTag), slog.Int("members", len(members)))
	if err != nil {
		s.log.Error("failed to load members from census", slog.String("outfit", outfitTag), sl.Err(err))
		return
	}
	wg.Add(1)
	go s.saveMembers(ctx, wg, outfitTag, members)
}

func (s *OutfitMembersSynchronizer) sync(ctx context.Context, wg *sync.WaitGroup) {
	outfits, err := s.outfitsLoader.Load(ctx)
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
