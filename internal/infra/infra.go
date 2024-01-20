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

func Op(op string) slog.Attr {
	return slog.String("op", op)
}

func WithOp(log *slog.Logger, op string) *slog.Logger {
	return log.With(Op(op))
}

func OpLogger(ctx context.Context, op string) *slog.Logger {
	return WithOp(Logger(ctx), op)
}
