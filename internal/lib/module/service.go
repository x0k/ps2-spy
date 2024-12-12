package module

import "context"

type StartFn func(ctx context.Context) error

type SimpleStartFn func(ctx context.Context)

type Service interface {
	Name() string
	Start(ctx context.Context) error
}

type starter struct {
	name  string
	start StartFn
}

func NewService(name string, start StartFn) Service {
	return &starter{name: name, start: start}
}

func (s *starter) Name() string {
	return s.name
}

func (s *starter) Start(ctx context.Context) error {
	return s.start(ctx)
}
