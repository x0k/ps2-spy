package stats_tracker

import (
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
)

type TaskId int64

type Task struct {
	Id              TaskId
	ChannelId       discord.ChannelId
	UtcStartWeekday time.Weekday
	UtcStartTime    time.Duration
	UtcEndWeekday   time.Weekday
	UtcEndTime      time.Duration
}

type ErrTaskDurationTooLong struct {
	MaxDuration time.Duration
	GotDuration time.Duration
}

func (e ErrTaskDurationTooLong) Error() string {
	return fmt.Sprintf(
		"stats tracker task duration too long: expected max %s, got %s",
		e.MaxDuration,
		e.GotDuration,
	)
}

type ErrMaxAmountOfTasksExceeded int

func (e ErrMaxAmountOfTasksExceeded) Error() string {
	return fmt.Sprintf("max amount of tasks exceeded: expected %d", e)
}

type ErrOverlappingTasks struct {
	Offset         time.Duration
	LocalWeekday   time.Weekday
	LocalStartTime time.Duration
	Duration       time.Duration
	Tasks          []Task
}

func (e ErrOverlappingTasks) Error() string {
	return fmt.Sprintf(
		"stats tracker task with weekday %d and start time %s intersects with %d existing tasks",
		e.LocalWeekday,
		e.LocalStartTime,
		len(e.Tasks),
	)
}
