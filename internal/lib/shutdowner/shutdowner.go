package shutdowner

import "context"

type Abstract interface {
	Shutdown(ctx context.Context) error
}

type noop struct{}

func (s noop) Shutdown(ctx context.Context) error {
	return nil
}

var Noop = noop{}
