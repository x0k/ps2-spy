package sql_storage

import (
	"context"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
)

func (s *Storage) StatsTrackerTasks(ctx context.Context, now time.Time) ([]discord.ChannelId, error) {
	utc := now.UTC()
	utcTime := utc.Hour()*int(time.Hour) + utc.Minute()*int(time.Minute)
	data, err := s.queries.ListActiveStatsTrackerTasks(ctx, db.ListActiveStatsTrackerTasksParams{
		Weekday: int64(utc.Weekday()),
		UtcTime: int64(utcTime),
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

func (s *Storage) ChannelStatsTrackerTasksLoader(
	ctx context.Context,
	channelId discord.ChannelId,
) ([]discord.StatsTrackerTask, error) {
	data, err := s.queries.ListChannelStatsTrackerTasks(ctx, string(channelId))
	if err != nil {
		return nil, fmt.Errorf("failed to list channel %q stats tracker tasks: %w", string(channelId), err)
	}
	tasks := make([]discord.StatsTrackerTask, 0, len(data))
	for _, t := range data {
		tasks = append(tasks, discord.StatsTrackerTask{
			Id:           discord.StatsTrackerTaskId(t.TaskID),
			ChannelId:    channelId,
			UtcWeekday:   time.Weekday(t.Weekday),
			UtcStartTime: time.Duration(t.UtcStartTime),
			UtcEndTime:   time.Duration(t.UtcEndTime),
		})
	}
	return tasks, nil
}
