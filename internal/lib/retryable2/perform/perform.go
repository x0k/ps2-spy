package perform

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

func Log(
	log *slog.Logger,
	lvl slog.Level,
	msg string,
	args ...slog.Attr,
) func(context.Context, error) {
	return func(ctx context.Context, err error) {
		log.LogAttrs(
			ctx,
			lvl,
			msg,
			append(args, sl.Err(err))...,
		)
	}
}

func ExponentialBackoff(
	d time.Duration,
) func(context.Context, error) {
	return func(ctx context.Context, err error) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
			d *= 2
		}
	}
}
