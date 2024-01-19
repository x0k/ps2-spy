package infra

import (
	"context"
	"log/slog"
	"sync"
)

const (
	LoggerKey = "log"
	WgKey     = "wg"
)

func Wg(ctx context.Context) *sync.WaitGroup {
	return ctx.Value(WgKey).(*sync.WaitGroup)
}

func Logger(ctx context.Context) *slog.Logger {
	return ctx.Value(LoggerKey).(*slog.Logger)
}

func OpLogger(ctx context.Context, op string) *slog.Logger {
	return Logger(ctx).With(slog.String("op", op))
}
