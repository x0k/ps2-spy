package stats_tracker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_loadout "github.com/x0k/ps2-spy/internal/ps2/loadout"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

type StatsTracker struct {
	mu                 sync.RWMutex
	wg                 sync.WaitGroup
	log                *logger.Logger
	channels           map[discord.ChannelId]*ChannelTracker
	publisher          pubsub.Publisher[Event]
	trackingManager    *tracking_manager.TrackingManager
	maxTracingDuration time.Duration
}

func New(
	log *logger.Logger,
	publisher pubsub.Publisher[Event],
	trackingManager *tracking_manager.TrackingManager,
	maxTrackingDuration time.Duration,
) *StatsTracker {
	return &StatsTracker{
		log:                log,
		channels:           make(map[discord.ChannelId]*ChannelTracker),
		publisher:          publisher,
		trackingManager:    trackingManager,
		maxTracingDuration: maxTrackingDuration,
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
			if err := s.handleTrackersOvertime(); err != nil {
				s.log.Error(ctx, "error during handleTrackersOvertime", sl.Err(err))
			}
		}
	}
}

func (s *StatsTracker) StartChannelTracker(channelId discord.ChannelId) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stopErr := s.stopChannelTracker(channelId)
	startErr := s.startChannelTracker(channelId)
	return errors.Join(stopErr, startErr)
}

func (s *StatsTracker) StopChannelTracker(channelId discord.ChannelId) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopChannelTracker(channelId)
}

func (s *StatsTracker) HandleDeathEvent(ctx context.Context, event events.Death) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.handleDeathEvent(ctx, event); err != nil {
			s.log.Error(ctx, "error during handleDeathEvent", sl.Err(err))
		}
	}()
}

func (s *StatsTracker) HandleGainExperienceEvent(ctx context.Context, event events.GainExperience) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.handleGainExperienceEvent(ctx, event); err != nil {
			s.log.Error(ctx, "error during handleGainExperienceEvent", sl.Err(err))
		}
	}()
}

func (s *StatsTracker) startChannelTracker(channelId discord.ChannelId) error {
	tracker := newChannelTracker()
	s.channels[channelId] = tracker
	return s.publisher.Publish(ChannelTrackerStarted{
		ChannelId: channelId,
		StartedAt: tracker.startedAt,
	})
}

func (s *StatsTracker) stopChannelTracker(channelId discord.ChannelId) error {
	tracker, ok := s.channels[channelId]
	if !ok {
		return nil
	}
	delete(s.channels, channelId)
	return s.publisher.Publish(ChannelTrackerStopped{
		ChannelId:  channelId,
		StartedAt:  tracker.startedAt,
		Characters: tracker.characters,
	})
}

func (s *StatsTracker) handleTrackersOvertime() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for channelId, tracker := range s.channels {
		if time.Since(tracker.startedAt) > s.maxTracingDuration {
			// NOTE: We are modifying the map while iterating over it
			if err := s.stopChannelTracker(channelId); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (s *StatsTracker) handleGainExperienceEvent(ctx context.Context, event events.GainExperience) error {
	charId := ps2.CharacterId(event.CharacterID)
	if channels, err := s.trackingManager.ChannelIdsForCharacter(ctx, charId); err == nil {
		s.handleCharacterEvent(
			ctx,
			channels,
			charId,
			ps2_loadout.Loadout(event.LoadoutID),
			updateLoadout,
		)
	} else {
		return fmt.Errorf("cannot get channels for character %q: %w", charId, err)
	}
	return nil
}

func (s *StatsTracker) handleDeathEvent(ctx context.Context, event events.Death) error {
	var errs []error
	charId := ps2.CharacterId(event.CharacterID)
	isSuicide := event.AttackerCharacterID == "0" || event.AttackerCharacterID == event.CharacterID
	deathAdder := addDeath
	if isSuicide {
		deathAdder = addSuicide
	}
	if channels, err := s.trackingManager.ChannelIdsForCharacter(ctx, charId); err == nil {
		s.handleCharacterEvent(
			ctx,
			channels,
			charId,
			ps2_loadout.Loadout(event.CharacterLoadoutID),
			deathAdder,
		)
	} else {
		errs = append(errs, fmt.Errorf("cannot get channels for character %q: %w", charId, err))
	}
	if isSuicide {
		return errors.Join(errs...)
	}
	charId = ps2.CharacterId(event.AttackerCharacterID)
	killAdder := addBodyKill
	if event.AttackerTeamID == event.TeamID {
		killAdder = addTeamKill
	} else if event.IsHeadshot == "1" {
		killAdder = addHeadShotKill
	}
	if channels, err := s.trackingManager.ChannelIdsForCharacter(ctx, charId); err == nil {
		s.handleCharacterEvent(
			ctx,
			channels,
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
	channels []discord.Channel,
	characterId ps2.CharacterId,
	loadout ps2_loadout.Loadout,
	update func(*CharacterStats),
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
	for _, channel := range channels {
		if tracker, ok := s.channels[channel.ChannelId]; ok {
			tracker.handleCharacterEvent(characterId, loadoutType, update)
		}
	}
}
