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
	log          *slog.Logger
	censusLoader loaders.KeyedLoader[string, []string]
	membersSaver Saver
	// Platform specific
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
		log:             log,
		censusLoader:    censusMembersLoader,
		outfitsLoader:   trackableOutfitsLoader,
		membersSaver:    membersSaver,
		refreshInterval: refreshInterval,
	}
}

func (s *OutfitMembersSynchronizer) saveMembers(ctx context.Context, wg *sync.WaitGroup, outfit string, members []string) {
	defer wg.Done()
	if err := s.membersSaver.Save(ctx, outfit, members); err != nil {
		s.log.Error("failed to save members", slog.String("outfit", outfit), sl.Err(err))
	}
}

func (s *OutfitMembersSynchronizer) sync(ctx context.Context, wg *sync.WaitGroup) {
	outfits, err := s.outfitsLoader.Load(ctx)
	if err != nil {
		s.log.Error("failed to load trackable outfits", sl.Err(err))
		return
	}
	for _, outfits := range outfits {
		select {
		case <-ctx.Done():
			return
		default:
			members, err := s.censusLoader.Load(ctx, outfits)
			if err != nil {
				s.log.Error("failed to load members from census", slog.String("outfit", outfits), sl.Err(err))
				continue
			}
			wg.Add(1)
			go s.saveMembers(ctx, wg, outfits, members)
		}
	}
}

func (s *OutfitMembersSynchronizer) Start(ctx context.Context, wg *sync.WaitGroup) {
	if s.started.Swap(true) {
		return
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
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
