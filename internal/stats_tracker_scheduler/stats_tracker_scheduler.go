package stats_tracker_scheduler

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type StatsTasksLoader = loader.Keyed[time.Time, []discord.ChannelId]

type StatsTrackerScheduler struct {
	log          *logger.Logger
	statsTracker *stats_tracker.StatsTracker
	tasksLoader  StatsTasksLoader

	tasks []discord.ChannelId
}

func New(
	log *logger.Logger,
	statsTracker *stats_tracker.StatsTracker,
	tasksLoader StatsTasksLoader,
) *StatsTrackerScheduler {
	return &StatsTrackerScheduler{
		log:          log,
		statsTracker: statsTracker,
		tasksLoader:  tasksLoader,
	}
}

func (s *StatsTrackerScheduler) Start(ctx context.Context) {
	t := time.NewTicker(1 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			s.invalidateStatsTrackers(ctx, now)
		}
	}
}

func (s *StatsTrackerScheduler) invalidateStatsTrackers(ctx context.Context, now time.Time) {
	newTasks, err := s.tasksLoader(ctx, now)
	if err != nil {
		s.log.Error(ctx, "error loading tasks", sl.Err(err))
		return
	}
	d := diff.SlicesDiff(s.tasks, newTasks)
	for _, channelId := range d.ToDel {
		if err := s.statsTracker.StopChannelTracker(ctx, channelId); err != nil {
			s.log.Error(ctx, "failed to stop channel tracker", sl.Err(err))
		}
	}
	for _, channelId := range d.ToAdd {
		if err := s.statsTracker.StartChannelTracker(ctx, channelId); err != nil {
			s.log.Error(ctx, "failed to start channel tracker", sl.Err(err))
		}
	}
	s.tasks = newTasks
}
