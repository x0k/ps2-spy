package taskx_test

import (
	"sync/atomic"
	"testing"

	"github.com/x0k/ps2-spy/internal/lib/taskx"
)

func TestGroupWaitsForRunningTasks(t *testing.T) {
	g := taskx.NewGroup()
	var done atomic.Bool
	g.Go(func() {
		done.Store(true)
	})
	g.Wait()
	if !done.Load() {
		t.Fatal("Wait returned before task finished")
	}
}

func TestGroupDoesNotStartTasksAfterClose(t *testing.T) {
	g := taskx.NewGroup()
	g.Close()
	var started atomic.Bool
	if g.Go(func() { started.Store(true) }) {
		t.Fatal("Go returned true after Close")
	}
	if started.Load() {
		t.Fatal("task started after Close")
	}
}

func TestGroupAllowsGoDuringWait(t *testing.T) {
	g := taskx.NewGroup()
	first := make(chan struct{})
	second := make(chan struct{})
	g.Go(func() {
		<-first
		g.Go(func() {
			close(second)
		})
	})
	close(first)
	g.Wait()
	<-second
}
