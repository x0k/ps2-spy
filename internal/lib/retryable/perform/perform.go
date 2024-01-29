package perform

import (
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
)

func Debug(log *slog.Logger, msg string) func(r *retryable.Retryable) {
	return func(r *retryable.Retryable) {
		if log == nil {
			return
		}
		log.Debug(msg, sl.Err(r.Err), slog.Duration("suspense_duration", r.SuspenseDuration))
	}
}

func RecoverSuspenseDuration(recovered time.Duration) func(r *retryable.Retryable) {
	startTime := time.Now()
	return func(r *retryable.Retryable) {
		now := time.Now()
		if now.Sub(startTime) > r.SuspenseDuration {
			r.SuspenseDuration = recovered
		}
		startTime = now.Add(r.SuspenseDuration)
	}
}
