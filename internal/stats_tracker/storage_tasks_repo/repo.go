package storage_stats_tracker_tasks_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
)

type Repository struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) GetByChannelId(
	ctx context.Context, channelId discord.ChannelId,
) ([]stats_tracker.Task, error) {
	data, err := r.storage.Queries().ListChannelStatsTrackerTasks(
		ctx, string(channelId),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list channel %q stats tracker tasks: %w", channelId, err)
	}
	tasks := make([]stats_tracker.Task, 0, len(data))
	for _, t := range data {
		tasks = append(tasks, statsTrackerTaskFromDTO(t))
	}
	return tasks, nil
}

func statsTrackerTaskFromDTO(task db.StatsTrackerTask) stats_tracker.Task {
	return stats_tracker.Task{
		Id:              stats_tracker.TaskId(task.TaskID),
		ChannelId:       discord.ChannelId(task.ChannelID),
		UtcStartWeekday: time.Weekday(task.UtcStartWeekday),
		UtcStartTime:    time.Duration(task.UtcStartTime),
		UtcEndWeekday:   time.Weekday(task.UtcEndWeekday),
		UtcEndTime:      time.Duration(task.UtcEndTime),
	}
}
