package stats_tracker_tasks_creator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type fakeTasksRepo struct {
	tasks      []stats_tracker.Task
	deleted    []stats_tracker.TaskId
	count      int64
	overlap    []stats_tracker.Task
	inTx       bool
	createErr  error
	deleteErr  error
	countErr   error
	overlapErr error
}

func (r *fakeTasksRepo) Create(ctx context.Context, task stats_tracker.Task) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.tasks = append(r.tasks, task)
	return nil
}

func (r *fakeTasksRepo) Delete(ctx context.Context, channelId discord.ChannelId, taskId stats_tracker.TaskId) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deleted = append(r.deleted, taskId)
	return nil
}

func (r *fakeTasksRepo) ChannelTasksCount(context.Context, discord.ChannelId) (int64, error) {
	return r.count, r.countErr
}

func (r *fakeTasksRepo) OverlappingTasks(context.Context, stats_tracker.Task) ([]stats_tracker.Task, error) {
	return r.overlap, r.overlapErr
}

func (r *fakeTasksRepo) Transaction(ctx context.Context, run func(r *fakeTasksRepo) error) error {
	r.inTx = true
	return run(r)
}

func TestCreateConvertsLocalTaskToUTC(t *testing.T) {
	repo := &fakeTasksRepo{}
	creator := New[*fakeTasksRepo](repo, 4*time.Hour, 7)
	loc := time.FixedZone("UTC+3", int((3 * time.Hour).Seconds()))

	err := creator.Create(context.Background(), stats_tracker.CreateOrUpdateTask{
		ChannelId:      "channel",
		Timezone:       loc,
		LocalWeekdays:  []time.Weekday{time.Monday},
		LocalStartHour: 1,
		LocalStartMin:  30,
		Duration:       2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !repo.inTx {
		t.Fatal("Create() did not run inside a transaction")
	}
	if len(repo.tasks) != 1 {
		t.Fatalf("created tasks = %#v", repo.tasks)
	}
	task := repo.tasks[0]
	if task.UtcStartWeekday != time.Sunday || task.UtcStartTime != 22*time.Hour+30*time.Minute {
		t.Fatalf("unexpected UTC start: weekday=%v time=%v", task.UtcStartWeekday, task.UtcStartTime)
	}
	if task.UtcEndWeekday != time.Monday || task.UtcEndTime != 30*time.Minute {
		t.Fatalf("unexpected UTC end: weekday=%v time=%v", task.UtcEndWeekday, task.UtcEndTime)
	}
}

func TestCreateRejectsTooManyTasks(t *testing.T) {
	repo := &fakeTasksRepo{count: 6}
	creator := New[*fakeTasksRepo](repo, 4*time.Hour, 7)

	err := creator.Create(context.Background(), stats_tracker.CreateOrUpdateTask{
		ChannelId:      "channel",
		Timezone:       time.UTC,
		LocalWeekdays:  []time.Weekday{time.Monday, time.Tuesday},
		LocalStartHour: 12,
		Duration:       time.Hour,
	})
	var tooMany stats_tracker.ErrTooManyTasksPerChannel
	if !errors.As(err, &tooMany) {
		t.Fatalf("Create() error = %v, want ErrTooManyTasksPerChannel", err)
	}
	if len(repo.tasks) != 0 {
		t.Fatalf("created tasks despite error: %#v", repo.tasks)
	}
}

func TestUpdateDeletesExistingTaskBeforeCreatingReplacement(t *testing.T) {
	repo := &fakeTasksRepo{}
	creator := New[*fakeTasksRepo](repo, 4*time.Hour, 7)

	err := creator.Update(context.Background(), stats_tracker.CreateOrUpdateTask{
		Id:             42,
		ChannelId:      "channel",
		Timezone:       time.UTC,
		LocalWeekdays:  []time.Weekday{time.Friday},
		LocalStartHour: 18,
		Duration:       time.Hour,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != 42 {
		t.Fatalf("deleted tasks = %#v", repo.deleted)
	}
	if len(repo.tasks) != 1 {
		t.Fatalf("created tasks = %#v", repo.tasks)
	}
}

func TestCreateRejectsOverlappingTasks(t *testing.T) {
	repo := &fakeTasksRepo{
		overlap: []stats_tracker.Task{{Id: 7, ChannelId: "channel"}},
	}
	creator := New[*fakeTasksRepo](repo, 4*time.Hour, 7)

	err := creator.Create(context.Background(), stats_tracker.CreateOrUpdateTask{
		ChannelId:      "channel",
		Timezone:       time.UTC,
		LocalWeekdays:  []time.Weekday{time.Monday},
		LocalStartHour: 12,
		Duration:       time.Hour,
	})
	var overlapping stats_tracker.ErrOverlappingTasks
	if !errors.As(err, &overlapping) {
		t.Fatalf("Create() error = %v, want ErrOverlappingTasks", err)
	}
	if len(repo.tasks) != 0 {
		t.Fatalf("created tasks despite overlap: %#v", repo.tasks)
	}
}
