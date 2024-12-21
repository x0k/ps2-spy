package sql_storage

import (
	"context"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/shared"
)

func (s *Storage) ActiveStatsTrackerTasks(ctx context.Context, now time.Time) ([]discord.ChannelId, error) {
	utc := now.UTC()
	utcTime := utc.Hour()*int(time.Hour) + utc.Minute()*int(time.Minute)
	data, err := s.queries.ListActiveStatsTrackerTasks(ctx, db.ListActiveStatsTrackerTasksParams{
		UtcWeekday: int64(utc.Weekday()),
		UtcTime:    int64(utcTime),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list active stats tracker tasks: %w", err)
	}
	var channelIds []discord.ChannelId
	for _, channelId := range data {
		channelIds = append(channelIds, discord.ChannelId(channelId))
	}
	return channelIds, nil
}

func (s *Storage) ChannelStatsTrackerTasks(
	ctx context.Context,
	channelId discord.ChannelId,
) ([]discord.StatsTrackerTask, error) {
	data, err := s.queries.ListChannelStatsTrackerTasks(ctx, string(channelId))
	if err != nil {
		return nil, fmt.Errorf("failed to list channel %q stats tracker tasks: %w", string(channelId), err)
	}
	tasks := make([]discord.StatsTrackerTask, 0, len(data))
	for _, t := range data {
		tasks = append(tasks, statsTrackerTaskFromDTO(t))
	}
	return tasks, nil
}

func (s *Storage) StatsTrackerTask(ctx context.Context, taskId discord.StatsTrackerTaskId) (discord.StatsTrackerTask, error) {
	data, err := s.queries.GetStatsTrackerTask(ctx, int64(taskId))
	if err != nil {
		return discord.StatsTrackerTask{}, fmt.Errorf("failed to get stats tracker task %d: %w", taskId, err)
	}
	return statsTrackerTaskFromDTO(data), nil
}

func (s *Storage) CreateStatsTrackerTask(
	ctx context.Context,
	channelId discord.ChannelId,
	task discord.StatsTrackerTaskState,
) error {
	_, offsetInSeconds := time.Now().In(task.Timezone).Zone()
	offset := -(time.Duration(offsetInSeconds) * time.Second)
	return s.Begin(ctx, 0, func(s *Storage) error {
		return createStatsTrackerTask(ctx, s, channelId, offset, task)
	})
}

func (s *Storage) RemoveStatsTrackerTask(ctx context.Context, channelId discord.ChannelId, taskId discord.StatsTrackerTaskId) error {
	return s.queries.RemoveChannelStatsTrackerTask(ctx, db.RemoveChannelStatsTrackerTaskParams{
		ChannelID: string(channelId),
		TaskID:    int64(taskId),
	})
}

func (s *Storage) UpdateStatsTrackerTask(
	ctx context.Context,
	channelId discord.ChannelId,
	task discord.StatsTrackerTaskState,
) error {
	_, offsetInSeconds := time.Now().In(task.Timezone).Zone()
	offset := -(time.Duration(offsetInSeconds) * time.Second)
	return s.Begin(ctx, 0, func(s *Storage) error {
		if err := s.queries.RemoveChannelStatsTrackerTask(ctx, db.RemoveChannelStatsTrackerTaskParams{
			ChannelID: string(channelId),
			TaskID:    int64(task.TaskId),
		}); err != nil {
			return fmt.Errorf("failed to remove stats tracker task: %w", err)
		}
		return createStatsTrackerTask(ctx, s, channelId, offset, task)
	})
}

func createStatsTrackerTask(
	ctx context.Context,
	s *Storage,
	channelId discord.ChannelId,
	offset time.Duration,
	task discord.StatsTrackerTaskState,
) error {
	if task.Duration > 4*time.Hour {
		return fmt.Errorf("duration must be less than 4 hours")
	}
	for _, localWeekday := range task.LocalWeekdays {
		localStart := time.Duration(task.LocalStartHour)*time.Hour + time.Duration(task.LocalStartMin)*time.Minute
		utcStartWeekday, utcStartTime := shared.ShiftDate(localWeekday, localStart, offset)
		utcEndWeekday, utcEndTime := shared.ShiftDate(localWeekday, localStart, offset+task.Duration)
		if tasks, err := s.queries.ListChannelIntersectingStatsTrackerTasks(ctx, db.ListChannelIntersectingStatsTrackerTasksParams{
			ChannelID:    string(channelId),
			StartWeekday: int64(utcStartWeekday),
			StartTime:    int64(utcStartTime),
			EndWeekday:   int64(utcEndWeekday),
			EndTime:      int64(utcEndTime),
		}); err != nil {
			return fmt.Errorf("failed to list intersecting stats tracker tasks: %w", err)
		} else if len(tasks) > 0 {
			return fmt.Errorf(
				"stats tracker task %v with weekday %d intersects with %d existing tasks",
				task, localWeekday, len(tasks),
			)
		}
		if err := s.queries.InsertChannelStatsTrackerTask(ctx, db.InsertChannelStatsTrackerTaskParams{
			ChannelID:       string(channelId),
			UtcStartWeekday: int64(utcStartWeekday),
			UtcStartTime:    int64(utcStartTime),
			UtcEndWeekday:   int64(utcEndWeekday),
			UtcEndTime:      int64(utcEndTime),
		}); err != nil {
			return fmt.Errorf("failed to create stats tracker task: %w", err)
		}
	}
	return nil
}

func statsTrackerTaskFromDTO(task db.StatsTrackerTask) discord.StatsTrackerTask {
	return discord.StatsTrackerTask{
		Id:              discord.StatsTrackerTaskId(task.TaskID),
		ChannelId:       discord.ChannelId(task.ChannelID),
		UtcStartWeekday: time.Weekday(task.UtcStartWeekday),
		UtcStartTime:    time.Duration(task.UtcStartTime),
		UtcEndWeekday:   time.Weekday(task.UtcEndWeekday),
		UtcEndTime:      time.Duration(task.UtcEndTime),
	}
}
