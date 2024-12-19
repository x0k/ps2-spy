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
