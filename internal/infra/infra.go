package infra

import (
	"context"
	"log/slog"
	"sync"
)

type ContextKey string

const (
	LoggerKey ContextKey = "log"
	WgKey     ContextKey = "wg"
)

func Wg(ctx context.Context) *sync.WaitGroup {
	return ctx.Value(WgKey).(*sync.WaitGroup)
}

func Op(op string) slog.Attr {
	return slog.String("op", op)
}
