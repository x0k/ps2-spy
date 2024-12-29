package stats_tracker_storage_tasks_repo

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

func (r *Repository) ById(ctx context.Context, taskId stats_tracker.TaskId) (stats_tracker.Task, error) {
	data, err := r.storage.Queries().GetStatsTrackerTask(ctx, int64(taskId))
	if err != nil {
		return stats_tracker.Task{}, fmt.Errorf("failed to get stats tracker task %d: %w", taskId, err)
	}
	return statsTrackerTaskFromDTO(data), nil
}

func (r *Repository) ByChannelId(
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

func (r *Repository) ChannelIdsWithActiveTasks(ctx context.Context, now time.Time) ([]discord.ChannelId, error) {
	utc := now.UTC()
	utcTime := utc.Hour()*int(time.Hour) + utc.Minute()*int(time.Minute)
	data, err := r.storage.Queries().ListActiveStatsTrackerTasks(ctx, db.ListActiveStatsTrackerTasksParams{
		UtcWeekday: int64(utc.Weekday()),
		UtcTime:    int64(utcTime),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list active stats tracker tasks: %w", err)
	}
	channelIds := make([]discord.ChannelId, 0, len(data))
	for _, channelId := range data {
		channelIds = append(channelIds, discord.ChannelId(channelId))
	}
	return channelIds, nil
}

func (r *Repository) Create(ctx context.Context, task stats_tracker.Task) error {
	return r.storage.Queries().InsertChannelStatsTrackerTask(ctx, db.InsertChannelStatsTrackerTaskParams{
		ChannelID:       string(task.ChannelId),
		UtcStartWeekday: int64(task.UtcStartWeekday),
		UtcStartTime:    int64(task.UtcStartTime),
		UtcEndWeekday:   int64(task.UtcEndWeekday),
		UtcEndTime:      int64(task.UtcEndTime),
	})
}

func (r *Repository) Delete(ctx context.Context, channelId discord.ChannelId, taskId stats_tracker.TaskId) error {
	return r.storage.Queries().RemoveChannelStatsTrackerTask(ctx, db.RemoveChannelStatsTrackerTaskParams{
		TaskID:    int64(taskId),
		ChannelID: string(channelId),
	})
}

func (r *Repository) ChannelTasksCount(ctx context.Context, channelId discord.ChannelId) (int64, error) {
	return r.storage.Queries().GetCountChannelStatsTrackerTasks(ctx, string(channelId))
}

func (r *Repository) OverlappingTasks(ctx context.Context, task stats_tracker.Task) ([]stats_tracker.Task, error) {
	data, err := r.storage.Queries().ListChannelIntersectingStatsTrackerTasks(ctx, db.ListChannelIntersectingStatsTrackerTasksParams{
		ChannelID:    string(task.ChannelId),
		StartWeekday: int64(task.UtcStartWeekday),
		StartTime:    int64(task.UtcStartTime),
		EndWeekday:   int64(task.UtcEndWeekday),
		EndTime:      int64(task.UtcEndTime),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list overlapping stats tracker tasks: %w", err)
	}
	tasks := make([]stats_tracker.Task, 0, len(data))
	for _, t := range data {
		tasks = append(tasks, statsTrackerTaskFromDTO(t))
	}
	return tasks, nil
}

func (r *Repository) Transaction(ctx context.Context, fn func(r *Repository) error) error {
	return r.storage.Transaction(ctx, func(s storage.Storage) error {
		return fn(New(s))
	})
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
