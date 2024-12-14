package module

import "context"

type Run func(ctx context.Context) error

type Runnable interface {
	Name() string
	Run(ctx context.Context) error
}

type runner struct {
	name string
	run  Run
}

func NewRun(name string, start Run) *runner {
	return &runner{name: name, run: start}
}

func (s *runner) Name() string {
	return s.name
}

func (s *runner) Run(ctx context.Context) error {
	return s.run(ctx)
}

type voidRun func(ctx context.Context)

func newVoidRun(name string, run voidRun) Runnable {
	return NewRun(name, func(ctx context.Context) error {
		run(ctx)
		return nil
	})
}
