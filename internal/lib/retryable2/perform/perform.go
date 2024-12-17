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

func ExponentialBackoff(duration time.Duration) func(context.Context, error) {
	start := time.Now()
	d := duration
	return func(ctx context.Context, err error) {
		// Recover suspense duration
		now := time.Now()
		if now.Sub(start) > d {
			d = duration
		}
		start = now.Add(d)
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
			d *= 2
		}
	}
}
