package taskx

import "sync"

// Group runs fire-and-forget tasks and allows waiting for them to finish.
//
// Unlike sync.WaitGroup, Go may be called concurrently with Wait and after
// Close: tasks submitted after Close are not started. This makes the group
// safe to use from callbacks that may fire while the owner is shutting down.
type Group struct {
	mu     sync.Mutex
	cond   *sync.Cond
	count  int
	closed bool
}

func NewGroup() *Group {
	g := &Group{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

// Go starts fn in a new goroutine. It returns false if the group is closed and
// the task was not started.
func (g *Group) Go(fn func()) bool {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return false
	}
	g.count++
	g.mu.Unlock()

	go func() {
		defer g.done()
		fn()
	}()
	return true
}

// Wait blocks until all currently running tasks finish. New tasks may still be
// started unless Close was called.
func (g *Group) Wait() {
	g.mu.Lock()
	defer g.mu.Unlock()
	for g.count > 0 {
		g.cond.Wait()
	}
}

// Close prevents new tasks from starting and waits for running tasks to finish.
// It is safe to call Close multiple times.
func (g *Group) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.closed = true
	for g.count > 0 {
		g.cond.Wait()
	}
}

func (g *Group) done() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.count--
	if g.count == 0 {
		g.cond.Broadcast()
	}
}
