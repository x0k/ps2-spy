package perform

import (
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
)

func Debug(log *slog.Logger, msg string, args ...any) func(r *retryable.Base) {
	return func(r *retryable.Base) {
		log.Debug(msg, sl.Err(r.Err), slog.Duration("suspense_duration", r.SuspenseDuration))
	}
}

func RecoverSuspenseDuration(recovered time.Duration) func(r *retryable.Base) {
	startTime := time.Now()
	return func(r *retryable.Base) {
		now := time.Now()
		if now.Sub(startTime) > r.SuspenseDuration {
			r.SuspenseDuration = recovered
		}
		startTime = now.Add(r.SuspenseDuration)
	}
}
