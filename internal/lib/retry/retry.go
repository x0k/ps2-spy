package retry

import "time"

func WithRetry(action func() error, initialRetryInterval time.Duration) {

}
