package stats_tracker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_loadout "github.com/x0k/ps2-spy/internal/ps2/loadout"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrNothingToTrack = errors.New("nothing to track")
var ErrNoChannelTrackerToStop = errors.New("no channel tracker to stop")

type TrackablePlatformsLoader = loader.Keyed[discord.ChannelId, []ps2_platforms.Platform]
type CharacterTrackingChannelsLoader = loader.Keyed[discord.PlatformQuery[ps2.CharacterId], []discord.ChannelId]

type ChannelPlatformsTracker struct {
	trackers  map[ps2_platforms.Platform]*platformTracker
	startedAt time.Time
}

type StatsTracker struct {
	mu                       sync.RWMutex
	wg                       sync.WaitGroup
	log                      *logger.Logger
	trackers                 map[discord.ChannelId]ChannelPlatformsTracker
	publisher                pubsub.Publisher[Event]
	channelsLoader           CharacterTrackingChannelsLoader
	maxTracingDuration       time.Duration
	trackablePlatformsLoader TrackablePlatformsLoader
	charactersLoaders        map[ps2_platforms.Platform]CharactersLoader
}

func New(
	log *logger.Logger,
	publisher pubsub.Publisher[Event],
	channelsLoader CharacterTrackingChannelsLoader,
	trackablePlatformsLoader TrackablePlatformsLoader,
	charactersLoaders map[ps2_platforms.Platform]CharactersLoader,
	maxTrackingDuration time.Duration,
) *StatsTracker {
	return &StatsTracker{
		log:                      log,
		trackers:                 make(map[discord.ChannelId]ChannelPlatformsTracker),
		publisher:                publisher,
		channelsLoader:           channelsLoader,
		charactersLoaders:        charactersLoaders,
		maxTracingDuration:       maxTrackingDuration,
		trackablePlatformsLoader: trackablePlatformsLoader,
	}
}

func (s *StatsTracker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.wg.Wait()
			return
		case <-ticker.C:
			if err := s.handleTrackersOvertime(ctx); err != nil {
				s.log.Error(ctx, "error during handleTrackersOvertime", sl.Err(err))
			}
		}
	}
}

func (s *StatsTracker) StartChannelTracker(ctx context.Context, channelId discord.ChannelId) error {
	trackablePlatforms, err := s.trackablePlatformsLoader(ctx, channelId)
	if err != nil {
		return err
	}
	if len(trackablePlatforms) == 0 {
		return ErrNothingToTrack
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	stopErr := s.stopChannelTracker(ctx, channelId)
	if errors.Is(stopErr, ErrNoChannelTrackerToStop) {
		stopErr = nil
	}
	startErr := s.startChannelTracker(channelId, trackablePlatforms)
	return errors.Join(stopErr, startErr)
}

func (s *StatsTracker) StopChannelTracker(ctx context.Context, channelId discord.ChannelId) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopChannelTracker(ctx, channelId)
}

func (s *StatsTracker) HandleDeathEvent(ctx context.Context, platform ps2_platforms.Platform, event events.Death) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.handleDeathEvent(ctx, platform, event); err != nil {
			s.log.Error(ctx, "error during handleDeathEvent", sl.Err(err))
		}
	}()
}

func (s *StatsTracker) HandleGainExperienceEvent(ctx context.Context, platform ps2_platforms.Platform, event events.GainExperience) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.handleGainExperienceEvent(ctx, platform, event); err != nil {
			s.log.Error(ctx, "error during handleGainExperienceEvent", sl.Err(err))
		}
	}()
}

func (s *StatsTracker) startChannelTracker(channelId discord.ChannelId, trackablePlatforms []ps2_platforms.Platform) error {
	now := time.Now()
	trackers := make(map[ps2_platforms.Platform]*platformTracker, len(trackablePlatforms))
	s.trackers[channelId] = ChannelPlatformsTracker{
		startedAt: now,
		trackers:  trackers,
	}
	for _, platform := range trackablePlatforms {
		trackers[platform] = newPlatformTracker(platform, s.charactersLoaders[platform])
	}
	return s.publisher.Publish(ChannelTrackerStarted{
		ChannelId: channelId,
		StartedAt: now,
	})
}

func (s *StatsTracker) stopChannelTracker(ctx context.Context, channelId discord.ChannelId) error {
	pt, ok := s.trackers[channelId]
	if !ok {
		return ErrNoChannelTrackerToStop
	}
	stoppedAt := time.Now()
	delete(s.trackers, channelId)
	stats := make(map[ps2_platforms.Platform]PlatformStats, len(pt.trackers))
	for platform, tracker := range pt.trackers {
		var err error
		stats[platform], err = tracker.toStats(ctx, stoppedAt)
		if err != nil {
			s.log.Warn(ctx, "failed to get stats for platform", slog.String("channel_id", string(channelId)), slog.String("platform", string(platform)), sl.Err(err))
		}
	}
	return s.publisher.Publish(ChannelTrackerStopped{
		ChannelId: channelId,
		StartedAt: pt.startedAt,
		StoppedAt: stoppedAt,
		Platforms: stats,
	})
}

func (s *StatsTracker) handleTrackersOvertime(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for channelId, tracker := range s.trackers {
		if time.Since(tracker.startedAt) > s.maxTracingDuration {
			// NOTE: We are modifying the map while iterating over it
			if err := s.stopChannelTracker(ctx, channelId); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (s *StatsTracker) handleGainExperienceEvent(ctx context.Context, platform ps2_platforms.Platform, event events.GainExperience) error {
	charId := ps2.CharacterId(event.CharacterID)
	if channels, err := s.channelsLoader(ctx, discord.PlatformQuery[ps2.CharacterId]{Platform: platform, Value: charId}); err == nil {
		s.handleCharacterEvent(
			ctx,
			channels,
			platform,
			charId,
			ps2_loadout.Loadout(event.LoadoutID),
			updateLoadout,
		)
	} else {
		return fmt.Errorf("cannot get channels for character %q: %w", charId, err)
	}
	return nil
}

func (s *StatsTracker) handleDeathEvent(ctx context.Context, platform ps2_platforms.Platform, event events.Death) error {
	var errs []error
	charId := ps2.CharacterId(event.CharacterID)
	isDeathByRestrictedArea := event.AttackerCharacterID == "0"
	isSuicide := event.AttackerCharacterID == event.CharacterID
	deathAdder := addDeath
	if isSuicide {
		deathAdder = addSuicide
	} else if isDeathByRestrictedArea {
		deathAdder = addDeathByRestrictedArea
	}
	if channels, err := s.channelsLoader(ctx, discord.PlatformQuery[ps2.CharacterId]{Platform: platform, Value: charId}); err == nil {
		s.handleCharacterEvent(
			ctx,
			channels,
			platform,
			charId,
			ps2_loadout.Loadout(event.CharacterLoadoutID),
			deathAdder,
		)
	} else {
		errs = append(errs, fmt.Errorf("cannot get channels for character %q: %w", charId, err))
	}
	if isSuicide || isDeathByRestrictedArea {
		return errors.Join(errs...)
	}
	charId = ps2.CharacterId(event.AttackerCharacterID)
	killAdder := addBodyKill
	if event.AttackerTeamID == event.TeamID {
		killAdder = addTeamKill
	} else if event.IsHeadshot == "1" {
		killAdder = addHeadShotKill
	}
	if channels, err := s.channelsLoader(ctx, discord.PlatformQuery[ps2.CharacterId]{Platform: platform, Value: charId}); err == nil {
		s.handleCharacterEvent(
			ctx,
			channels,
			platform,
			charId,
			ps2_loadout.Loadout(event.AttackerLoadoutID),
			killAdder,
		)
	} else {
		errs = append(errs, fmt.Errorf("cannot get channels for character %q: %w", charId, err))
	}
	return errors.Join(errs...)
}

func (s *StatsTracker) handleCharacterEvent(
	ctx context.Context,
	channels []discord.ChannelId,
	platform ps2_platforms.Platform,
	characterId ps2.CharacterId,
	loadout ps2_loadout.Loadout,
	update func(*characterTracker),
) {
	if len(channels) == 0 {
		return
	}
	loadoutType, err := ps2_loadout.GetType(loadout)
	if err != nil {
		s.log.Warn(ctx, "cannot get loadout type, falling back to heavy assault", sl.Err(err))
		loadoutType = ps2_loadout.HeavyAssault
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, channelId := range channels {
		if platformsTracker, ok := s.trackers[channelId]; ok {
			if tracker, ok := platformsTracker.trackers[platform]; ok {
				tracker.handleCharacterEvent(characterId, loadoutType, update)
			}
		}
	}
}
