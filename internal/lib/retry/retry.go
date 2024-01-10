package retry

import (
	"context"
	"errors"
	"time"
)

var ErrMaxRetriesExceeded = errors.New("max retries exceeded")

type Retryable struct {
	Try               func() error
	While             func(err error, retryCount int) bool
	BaseRetryInterval time.Duration
	BeforeSleep       func(d time.Duration)
}

// If `Try` is executed longer than `BaseRetryInterval * 2^RetryCount`,
// then retries count will be recovered.
func RetryWhileWithRecover(rt Retryable) error {
	if rt.BaseRetryInterval == 0 {
		rt.BaseRetryInterval = time.Second
	}
	if rt.BeforeSleep == nil {
		rt.BeforeSleep = func(sleepInterval time.Duration) {}
	}
	shouldRetry := true
	retryInterval := rt.BaseRetryInterval
	backoffFactor := time.Duration(2)
	retryCount := 0
	for shouldRetry {
		startTime := time.Now()
		err := rt.Try()
		shouldRetry = rt.While(err, retryCount)
		if shouldRetry {
			if time.Since(startTime) > retryInterval {
				retryInterval = rt.BaseRetryInterval
				backoffFactor = time.Duration(2)
				retryCount = 0
			}
			rt.BeforeSleep(retryInterval)
			time.Sleep(retryInterval)
			retryInterval = backoffFactor * rt.BaseRetryInterval
			backoffFactor *= 2
			retryCount++
		} else {
			return err
		}
	}
	// unreachable
	return nil
}

func ContextIsNotCanceled(err error, _ int) bool {
	return !errors.Is(err, context.Canceled)
}
