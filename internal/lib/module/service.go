package module

import "context"

type Service interface {
	Name() string
	Start(ctx context.Context) error
}

type starter struct {
	name  string
	start func(ctx context.Context) error
}

func NewService(name string, start func(ctx context.Context) error) Service {
	return &starter{name: name, start: start}
}

func (s *starter) Name() string {
	return s.name
}

func (s *starter) Start(ctx context.Context) error {
	return s.start(ctx)
}
