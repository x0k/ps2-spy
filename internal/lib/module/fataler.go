package module

import (
	"context"
	"log/slog"
)

type Fataler interface {
	Fatal(ctx context.Context, err error)
}

func (m *Module) Fatal(ctx context.Context, err error) {
	if m.stopped.Swap(true) {
		m.log.LogAttrs(ctx, slog.LevelError, "fatal error", slog.String("error", err.Error()))
		return
	}
	m.fatal <- err
	close(m.fatal)
}
