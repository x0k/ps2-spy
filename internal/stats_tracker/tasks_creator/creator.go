package stats_tracker_tasks_creator

import (
	"context"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/timex"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type TasksRepoApi interface {
	Create(context.Context, stats_tracker.Task) error
	Delete(context.Context, discord.ChannelId, stats_tracker.TaskId) error
	ChannelTasksCount(context.Context, discord.ChannelId) (int64, error)
	OverlappingTasks(context.Context, stats_tracker.Task) ([]stats_tracker.Task, error)
}

type TasksRepo[R TasksRepoApi] interface {
	TasksRepoApi
	Transaction(context.Context, func(r R) error) error
}

type TasksCreator[R TasksRepoApi] struct {
	repo                       TasksRepo[R]
	maxTrackingDuration        time.Duration
	maxNumberOfTasksPerChannel int
}

func New[R TasksRepoApi](
	repo TasksRepo[R],
	maxTrackingDuration time.Duration,
	maxNumberOfTasksPerChannel int,
) *TasksCreator[R] {
	return &TasksCreator[R]{
		repo:                       repo,
		maxTrackingDuration:        maxTrackingDuration,
		maxNumberOfTasksPerChannel: maxNumberOfTasksPerChannel,
	}
}

func (c *TasksCreator[R]) Create(ctx context.Context, task stats_tracker.CreateOrUpdateTask) error {
	return c.repo.Transaction(ctx, func(r R) error {
		return createTask(ctx, c, r, task)
	})
}

func (c *TasksCreator[R]) Update(ctx context.Context, task stats_tracker.CreateOrUpdateTask) error {
	return c.repo.Transaction(ctx, func(r R) error {
		if err := r.Delete(ctx, task.ChannelId, task.Id); err != nil {
			return fmt.Errorf("failed to delete stats tracker task %d: %w", task.Id, err)
		}
		return createTask(ctx, c, r, task)
	})
}

func createTask[R TasksRepoApi](
	ctx context.Context,
	c *TasksCreator[R],
	r R,
	task stats_tracker.CreateOrUpdateTask,
) error {
	if task.Duration > c.maxTrackingDuration {
		return stats_tracker.ErrTaskDurationTooLong{
			MaxDuration: c.maxTrackingDuration,
			GotDuration: task.Duration,
		}
	}
	count, err := r.ChannelTasksCount(ctx, task.ChannelId)
	if err != nil {
		return fmt.Errorf("failed to get count of channel %q stats tracker tasks: %w", string(task.ChannelId), err)
	}
	finalCount := int(count) + len(task.LocalWeekdays)
	if finalCount > c.maxNumberOfTasksPerChannel {
		return stats_tracker.ErrMaxTooManyTasksPerChannel{
			Max: c.maxNumberOfTasksPerChannel,
			Got: finalCount,
		}
	}
	utcOffset := timex.LocationToOffset(task.Timezone)
	for _, localWeekday := range task.LocalWeekdays {
		utcStart := time.Duration(task.LocalStartHour)*time.Hour + time.Duration(task.LocalStartMin)*time.Minute - utcOffset
		utcStartWeekday, utcStartTime := timex.NormalizeDate(localWeekday, utcStart)
		utcEndWeekday, utcEndTime := timex.NormalizeDate(localWeekday, utcStart+task.Duration)
		t := stats_tracker.Task{
			ChannelId:       task.ChannelId,
			UtcStartWeekday: utcStartWeekday,
			UtcStartTime:    utcStartTime,
			UtcEndWeekday:   utcEndWeekday,
			UtcEndTime:      utcEndTime,
		}
		if tasks, err := r.OverlappingTasks(ctx, t); err != nil {
			return fmt.Errorf("failed to get overlapping tasks: %w", err)
		} else if len(tasks) > 0 {
			return stats_tracker.ErrOverlappingTasks{
				Timezone:       task.Timezone,
				LocalWeekday:   localWeekday,
				LocalStartTime: utcStart + utcOffset,
				Duration:       task.Duration,
				Tasks:          tasks,
			}
		}
		if err := r.Create(ctx, t); err != nil {
			return fmt.Errorf("failed to create stats tracker task: %w", err)
		}
	}
	return nil
}
