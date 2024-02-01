package perform

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
)

func Log(log *slog.Logger, lvl slog.Level, msg string, args ...slog.Attr) func(context.Context, *retryable.Retryable) {
	return func(ctx context.Context, r *retryable.Retryable) {
		log.LogAttrs(
			ctx,
			lvl,
			msg,
			append(args, sl.Err(r.Err), slog.Duration("suspense_duration", r.SuspenseDuration))...,
		)
	}
}

func RecoverSuspenseDuration(recovered time.Duration) func(context.Context, *retryable.Retryable) {
	startTime := time.Now()
	return func(_ context.Context, r *retryable.Retryable) {
		now := time.Now()
		if now.Sub(startTime) > r.SuspenseDuration {
			r.SuspenseDuration = recovered
		}
		startTime = now.Add(r.SuspenseDuration)
	}
}
