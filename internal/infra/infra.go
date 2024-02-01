package infra

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/logger"
)

const (
	LoggerKey = "log"
	WgKey     = "wg"
)

func Wg(ctx context.Context) *sync.WaitGroup {
	return ctx.Value(WgKey).(*sync.WaitGroup)
}

func Logger(ctx context.Context) *logger.Logger {
	return ctx.Value(LoggerKey).(*logger.Logger)
}

func Op(op string) slog.Attr {
	return slog.String("op", op)
}

func WithOp(log *logger.Logger, op string) *logger.Logger {
	return log.With(Op(op))
}

func OpLogger(ctx context.Context, op string) *logger.Logger {
	return WithOp(Logger(ctx), op)
}
