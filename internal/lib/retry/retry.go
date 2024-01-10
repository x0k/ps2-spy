package retry

import (
	"time"
)

type Retryable struct {
	Action            func() error
	Condition         func(error) bool
	BaseRetryInterval time.Duration
	BeforeSleep       func(d time.Duration)
}

func RetryWhile(rt Retryable) error {
	if rt.BaseRetryInterval == 0 {
		rt.BaseRetryInterval = time.Second
	}
	if rt.BeforeSleep == nil {
		rt.BeforeSleep = func(sleepInterval time.Duration) {}
	}
	shouldRetry := true
	retryInterval := rt.BaseRetryInterval
	backoffFactor := time.Duration(2)
	for shouldRetry {
		startTime := time.Now()
		err := rt.Action()
		shouldRetry = rt.Condition(err)
		if shouldRetry {
			if time.Since(startTime) > retryInterval {
				retryInterval = rt.BaseRetryInterval
				backoffFactor = time.Duration(2)
			}
			rt.BeforeSleep(retryInterval)
			time.Sleep(retryInterval)
			retryInterval = backoffFactor * rt.BaseRetryInterval
			backoffFactor *= 2
		} else {
			return err
		}
	}
	// unreachable
	return nil
}
