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
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/diff"
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
var ErrChannelStatsTrackerIsAlreadyStarted = errors.New("channel stats tracker is already started")

type ChannelTrackingPlatformsLoader = loader.Keyed[discord.ChannelId, []ps2_platforms.Platform]
type ChannelsWithActiveTasksLoader = loader.Keyed[time.Time, []discord.ChannelId]
type CharacterTrackingChannelsLoader = func(
	context.Context, ps2_platforms.Platform, ps2.CharacterId,
) ([]discord.ChannelId, error)

type StatsTracker struct {
	trackersMu          sync.RWMutex
	wg                  sync.WaitGroup
	log                 *logger.Logger
	trackers            map[discord.ChannelId]channelTracker
	publisher           pubsub.Publisher[Event]
	maxTrackingDuration time.Duration

	channelTrackingPlatformsLoader  ChannelTrackingPlatformsLoader
	channelWithActiveTasksLoader    ChannelsWithActiveTasksLoader
	characterTrackingChannelsLoader CharacterTrackingChannelsLoader
	charactersLoader                CharactersLoader

	scheduledTrackers []discord.ChannelId
	forceMu           sync.RWMutex
	forceStarted      *containers.ExpirationQueue[discord.ChannelId]
	forceStopped      *containers.ExpirationQueue[discord.ChannelId]
}

func New(
	log *logger.Logger,
	publisher pubsub.Publisher[Event],
	channelTrackingPlatformsLoader ChannelTrackingPlatformsLoader,
	channelWithActiveTasksLoader ChannelsWithActiveTasksLoader,
	characterTrackingChannelsLoader CharacterTrackingChannelsLoader,
	charactersLoader CharactersLoader,
	maxTrackingDuration time.Duration,
) *StatsTracker {
	return &StatsTracker{
		log:                             log,
		trackers:                        make(map[discord.ChannelId]channelTracker),
		publisher:                       publisher,
		maxTrackingDuration:             maxTrackingDuration,
		channelTrackingPlatformsLoader:  channelTrackingPlatformsLoader,
		channelWithActiveTasksLoader:    channelWithActiveTasksLoader,
		characterTrackingChannelsLoader: characterTrackingChannelsLoader,
		charactersLoader:                charactersLoader,
		forceStarted:                    containers.NewExpirationQueue[discord.ChannelId](),
		forceStopped:                    containers.NewExpirationQueue[discord.ChannelId](),
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
		case now := <-ticker.C:
			if err := s.handleTrackersOvertime(ctx); err != nil {
				s.log.Error(ctx, "error during handleTrackersOvertime", sl.Err(err))
			}
			s.invalidateForceTrackers(now)
			s.invalidateStatsTrackers(ctx, now)
		}
	}
}

func (s *StatsTracker) StartChannelTracker(ctx context.Context, channelId discord.ChannelId) error {
	return s.startChannelTracker(ctx, channelId, true)
}

func (s *StatsTracker) StopChannelTracker(ctx context.Context, channelId discord.ChannelId) error {
	return s.stopChannelTracker(ctx, channelId, true)
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

func (s *StatsTracker) startChannelTracker(ctx context.Context, channelId discord.ChannelId, force bool) error {
	// Started manually before scheduled task
	if !force && s.isForceStopped(channelId) {
		return nil
	}
	if s.isRunning(channelId) {
		return ErrChannelStatsTrackerIsAlreadyStarted
	}
	trackablePlatforms, err := s.channelTrackingPlatformsLoader(ctx, channelId)
	if err != nil {
		return err
	}
	if len(trackablePlatforms) == 0 {
		return ErrNothingToTrack
	}
	trackers := make(map[ps2_platforms.Platform]*platformTracker, len(trackablePlatforms))
	for _, platform := range trackablePlatforms {
		trackers[platform] = newPlatformTracker(platform, s.charactersLoader)
	}
	now := time.Now()
	s.trackersMu.Lock()
	s.trackers[channelId] = channelTracker{
		startedAt: now,
		trackers:  trackers,
	}
	s.trackersMu.Unlock()
	s.addStarted(channelId, force)
	s.publisher.Publish(ChannelTrackerStarted{
		ChannelId: channelId,
		StartedAt: now,
	})
	return nil
}

func (s *StatsTracker) stopChannelTracker(ctx context.Context, channelId discord.ChannelId, force bool) error {
	if !force && s.isForceStarted(channelId) {
		return nil
	}
	s.addStopped(channelId, force)
	pt, ok := s.popTracker(channelId)
	if !ok {
		return ErrNoChannelTrackerToStop
	}
	stoppedAt := time.Now()
	stats := make(map[ps2_platforms.Platform]PlatformStats, len(pt.trackers))
	for platform, tracker := range pt.trackers {
		var err error
		stats[platform], err = tracker.toStats(ctx, stoppedAt)
		if err != nil {
			s.log.Warn(ctx, "failed to get stats for platform", slog.String("channel_id", string(channelId)), slog.String("platform", string(platform)), sl.Err(err))
		}
	}
	s.publisher.Publish(ChannelTrackerStopped{
		ChannelId: channelId,
		StartedAt: pt.startedAt,
		StoppedAt: stoppedAt,
		Platforms: stats,
	})
	return nil
}

func (s *StatsTracker) popTracker(channelId discord.ChannelId) (channelTracker, bool) {
	s.trackersMu.Lock()
	defer s.trackersMu.Unlock()
	t, ok := s.trackers[channelId]
	if ok {
		delete(s.trackers, channelId)
		return t, true
	}
	return channelTracker{}, false
}

func (s *StatsTracker) handleTrackersOvertime(ctx context.Context) error {
	s.trackersMu.RLock()
	toStop := make([]discord.ChannelId, 0, len(s.trackers))
	for channelId, tracker := range s.trackers {
		if time.Since(tracker.startedAt) > s.maxTrackingDuration {
			toStop = append(toStop, channelId)
		}
	}
	s.trackersMu.RUnlock()
	var errs []error
	for _, channelId := range toStop {
		if err := s.stopChannelTracker(ctx, channelId, true); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *StatsTracker) handleGainExperienceEvent(ctx context.Context, platform ps2_platforms.Platform, event events.GainExperience) error {
	charId := ps2.CharacterId(event.CharacterID)
	if channels, err := s.characterTrackingChannelsLoader(ctx, platform, charId); err == nil {
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
	if channels, err := s.characterTrackingChannelsLoader(ctx, platform, charId); err == nil {
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
	if channels, err := s.characterTrackingChannelsLoader(ctx, platform, charId); err == nil {
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
	s.trackersMu.RLock()
	defer s.trackersMu.RUnlock()
	for _, channelId := range channels {
		if platformsTracker, ok := s.trackers[channelId]; ok {
			if tracker, ok := platformsTracker.trackers[platform]; ok {
				tracker.handleCharacterEvent(characterId, loadoutType, update)
			}
		}
	}
}

func (s *StatsTracker) invalidateStatsTrackers(ctx context.Context, now time.Time) {
	newTasks, err := s.channelWithActiveTasksLoader(ctx, now)
	if err != nil {
		s.log.Error(ctx, "error loading tasks", sl.Err(err))
		return
	}
	d := diff.SlicesDiff(s.scheduledTrackers, newTasks)
	for _, channelId := range d.ToDel {
		if err := s.stopChannelTracker(ctx, channelId, false); err != nil {
			s.log.Error(ctx, "failed to stop channel tracker", sl.Err(err))
		}
	}
	for _, channelId := range d.ToAdd {
		if err := s.startChannelTracker(ctx, channelId, false); err != nil {
			s.log.Error(ctx, "failed to start channel tracker", sl.Err(err))
		}
	}
	s.scheduledTrackers = newTasks
}

func (s *StatsTracker) isRunning(channelId discord.ChannelId) bool {
	s.trackersMu.RLock()
	defer s.trackersMu.RUnlock()
	_, ok := s.trackers[channelId]
	return ok
}

func (s *StatsTracker) isForceStarted(channelId discord.ChannelId) bool {
	s.forceMu.RLock()
	defer s.forceMu.RUnlock()
	return s.forceStarted.Has(channelId)
}

func (s *StatsTracker) addStarted(channelId discord.ChannelId, force bool) {
	s.forceMu.Lock()
	defer s.forceMu.Unlock()
	if force {
		s.forceStarted.Push(channelId)
	} else {
		s.forceStarted.Remove(channelId)
	}
	s.forceStopped.Remove(channelId)
}

func (s *StatsTracker) isForceStopped(channelId discord.ChannelId) bool {
	s.forceMu.RLock()
	defer s.forceMu.RUnlock()
	return s.forceStopped.Has(channelId)
}

func (s *StatsTracker) addStopped(channelId discord.ChannelId, force bool) {
	s.forceMu.Lock()
	defer s.forceMu.Unlock()
	if force {
		s.forceStopped.Push(channelId)
	} else {
		s.forceStopped.Remove(channelId)
	}
	s.forceStarted.Remove(channelId)
}

func (s *StatsTracker) invalidateForceTrackers(now time.Time) {
	s.forceMu.Lock()
	defer s.forceMu.Unlock()
	s.forceStarted.RemoveExpired(now.Add(-s.maxTrackingDuration), func(channelId discord.ChannelId) {})
	s.forceStopped.RemoveExpired(now.Add(-s.maxTrackingDuration), func(channelId discord.ChannelId) {})
}
