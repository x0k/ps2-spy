package while

import (
	"context"
	"errors"

	"github.com/x0k/ps2-spy/internal/lib/retryable"
)

func ErrorIsHere(r *retryable.Retryable) bool {
	return r.Err != nil
}

func RetryCountIsLessThan(maxRetries int) func(r *retryable.Retryable) bool {
	retryCount := -1
	return func(r *retryable.Retryable) bool {
		retryCount++
		return retryCount < maxRetries
	}
}

func ContextIsNotCancelled(r *retryable.Retryable) bool {
	return !errors.Is(r.Err, context.Canceled)
}
