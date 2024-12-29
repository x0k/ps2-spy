package stats_tracker

import (
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/timex"
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

type CreateOrUpdateTask struct {
	Id             TaskId
	ChannelId      discord.ChannelId
	Timezone       *time.Location
	LocalWeekdays  []time.Weekday
	LocalStartHour int
	LocalStartMin  int
	Duration       time.Duration
}

func NewCreateTask(
	timezone *time.Location,
) CreateOrUpdateTask {
	localNow := time.Now().In(timezone)
	localStartTime := time.Duration(localNow.Hour())*time.Hour + time.Duration(localNow.Minute()/10)*10*time.Minute
	localHour := int(localStartTime / time.Hour)
	localMin := int((localStartTime % time.Hour) / time.Minute)
	return CreateOrUpdateTask{
		Timezone:       timezone,
		LocalWeekdays:  []time.Weekday{localNow.Weekday()},
		LocalStartHour: localHour,
		LocalStartMin:  localMin,
		Duration:       2 * time.Hour,
	}
}

func NewUpdateTask(
	task Task,
	timezone *time.Location,
) CreateOrUpdateTask {
	utcOffset := timex.LocationToOffset(timezone)
	startWeekday, startTime := timex.NormalizeDate(task.UtcStartWeekday, task.UtcStartTime+utcOffset)
	endWeekday, endTime := timex.NormalizeDate(task.UtcEndWeekday, task.UtcEndTime+utcOffset)
	duration := endTime - startTime
	if startWeekday != endWeekday {
		duration += 24 * time.Hour
	}
	return CreateOrUpdateTask{
		Id:             task.Id,
		ChannelId:      task.ChannelId,
		Timezone:       timezone,
		LocalWeekdays:  []time.Weekday{startWeekday},
		LocalStartHour: int(startTime / time.Hour),
		LocalStartMin:  int((startTime % time.Hour) / time.Minute),
		Duration:       duration,
	}
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

type ErrMaxTooManyTasksPerChannel struct {
	Max int
	Got int
}

func (e ErrMaxTooManyTasksPerChannel) Error() string {
	return fmt.Sprintf("max amount of tasks per channel is %d, got %d", e.Max, e.Got)
}

type ErrOverlappingTasks struct {
	Timezone       *time.Location
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
