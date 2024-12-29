package ps2_outfit_members_synchronizer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
)

type ExternalOutfitsRepo interface {
	ActualMemberIds(context.Context, ps2_platforms.Platform, ps2.OutfitId) ([]ps2.CharacterId, error)
}

type OutfitsRepoApi interface {
	SynchronizedAt(context.Context, ps2_platforms.Platform, ps2.OutfitId) (time.Time, error)
	SaveSynchronizedAt(context.Context, ps2_platforms.Platform, ps2.OutfitId, time.Time) error
	TrackableOutfitIds(context.Context, ps2_platforms.Platform) ([]ps2.OutfitId, error)

	MemberIds(context.Context, ps2_platforms.Platform, ps2.OutfitId) ([]ps2.CharacterId, error)
	AddMember(context.Context, ps2_platforms.Platform, ps2.OutfitId, ps2.CharacterId) error
	RemoveMembers(context.Context, ps2_platforms.Platform, ps2.OutfitId, []ps2.CharacterId) error
}

type OutfitsRepo[R OutfitsRepoApi] interface {
	OutfitsRepoApi
	Transaction(ctx context.Context, run func(s R) error) error
}

type OutfitMembersSynchronizer[R OutfitsRepoApi] struct {
	log *logger.Logger
	wg  sync.WaitGroup

	externalRepo ExternalOutfitsRepo
	repo         OutfitsRepo[R]
	publisher    pubsub.Publisher[ps2.Event]

	refreshInterval time.Duration
}

func New[R OutfitsRepoApi](
	log *logger.Logger,
	repo OutfitsRepo[R],
	externalRepo ExternalOutfitsRepo,
	refreshInterval time.Duration,
	publisher pubsub.Publisher[ps2.Event],
) *OutfitMembersSynchronizer[R] {
	return &OutfitMembersSynchronizer[R]{
		log:             log,
		repo:            repo,
		externalRepo:    externalRepo,
		refreshInterval: refreshInterval,
		publisher:       publisher,
	}
}

func (s *OutfitMembersSynchronizer[R]) Start(ctx context.Context) {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()
	s.syncOutfits(ctx, time.Now())
	for {
		select {
		case <-ctx.Done():
			s.wg.Wait()
			return
		case now := <-ticker.C:
			s.syncOutfits(ctx, now)
		}
	}
}

func (s *OutfitMembersSynchronizer[R]) SyncOutfit(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.syncOutfit(ctx, platform, outfitId, time.Now())
	}()
}

func (s *OutfitMembersSynchronizer[R]) syncOutfits(ctx context.Context, now time.Time) {
	for _, platform := range ps2_platforms.Platforms {
		select {
		case <-ctx.Done():
			return
		default:
			s.syncPlatformOutfits(ctx, platform, now)
		}
	}
}

func (s *OutfitMembersSynchronizer[R]) syncPlatformOutfits(
	ctx context.Context, platform ps2_platforms.Platform, now time.Time,
) {
	outfits, err := s.repo.TrackableOutfitIds(ctx, platform)
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
			s.syncOutfit(ctx, platform, outfit, now)
		}
	}
}

func (s *OutfitMembersSynchronizer[R]) syncOutfit(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, now time.Time,
) {
	log := s.log.With(slog.String("outfit_id", string(outfitId)))
	err := retryable.New(func(ctx context.Context) error {
		syncAt, err := s.repo.SynchronizedAt(ctx, ps2_platforms.Platform(platform), outfitId)
		isNotFound := errors.Is(err, shared.ErrNotFound)
		if err != nil && !isNotFound {
			return fmt.Errorf("failed to load last sync time: %w", err)
		}
		if !isNotFound && time.Since(syncAt) < s.refreshInterval {
			log.Debug(ctx, "skipping sync")
			return nil
		}
		members, err := s.externalRepo.ActualMemberIds(ctx, platform, outfitId)
		log.Debug(ctx, "synchronizing", slog.Int("members", len(members)))
		if err != nil {
			return fmt.Errorf("failed to load members from census: %w", err)
		}
		if err := s.updateMembers(ctx, platform, outfitId, members, now); err != nil {
			return fmt.Errorf("failed to save members: %w", err)
		}
		return nil
	})(
		ctx,
		while.ErrorIsHere,
		while.ContextIsNotCancelled,
		while.HasAttempts(3),
		perform.Log(s.log.Logger, slog.LevelDebug, "members sync failed, retrying"),
		perform.ExponentialBackoff(1*time.Second),
	)
	if err != nil {
		log.Error(ctx, "failed to sync", sl.Err(err))
	}
}

func (u *OutfitMembersSynchronizer[R]) updateMembers(
	ctx context.Context,
	platform ps2_platforms.Platform,
	outfitId ps2.OutfitId,
	newMembers []ps2.CharacterId,
	now time.Time,
) error {
	var oldMembers []ps2.CharacterId
	var membersDiff diff.Diff[ps2.CharacterId]
	if err := u.repo.Transaction(ctx, func(r R) error {
		var err error
		if oldMembers, err = r.MemberIds(ctx, platform, outfitId); err != nil {
			return fmt.Errorf("failed to list members: %w", err)
		}
		membersDiff = diff.SlicesDiff(oldMembers, newMembers)
		if membersDiff.IsEmpty() {
			return nil
		}
		if err := r.RemoveMembers(ctx, platform, outfitId, membersDiff.ToDel); err != nil {
			return fmt.Errorf("failed to remove members: %w", err)
		}
		for _, member := range membersDiff.ToAdd {
			if err := r.AddMember(ctx, platform, outfitId, member); err != nil {
				return fmt.Errorf("failed to add member: %w", err)
			}
		}
		if err := r.SaveSynchronizedAt(ctx, platform, outfitId, now); err != nil {
			return fmt.Errorf("failed to save sync time: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	if len(membersDiff.ToAdd) > 0 {
		u.publisher.Publish(ps2.OutfitMembersAdded{
			Platform:     platform,
			OutfitId:     outfitId,
			CharacterIds: membersDiff.ToAdd,
		})
	}
	if len(membersDiff.ToDel) > 0 {
		u.publisher.Publish(ps2.OutfitMembersRemoved{
			Platform:     platform,
			OutfitId:     outfitId,
			CharacterIds: membersDiff.ToDel,
		})
	}
	if len(oldMembers) > 0 {
		u.publisher.Publish(ps2.OutfitMembersUpdate{
			Platform: platform,
			OutfitId: outfitId,
			Members:  membersDiff,
		})
	}
	return nil
}
