package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/retryable2"
	"github.com/x0k/ps2-spy/internal/lib/retryable2/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable2/while"
)

type RetryRoundTripper struct {
	timeout time.Duration
	trip    func(context.Context, *http.Request) (*http.Response, error)

	mu       sync.Mutex
	cancelId int64
	cancels  map[int64]func()
}

func NewRetryRoundTripper(
	log *slog.Logger,
	timeout time.Duration,
	roundTripper http.RoundTripper,
) *RetryRoundTripper {
	trip := retryable2.NewWithArg(
		func(ctx context.Context, req *http.Request) (*http.Response, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return roundTripper.RoundTrip(req.WithContext(ctx))
		},
		isRetryable,
		while.ContextIsNotCancelled,
		while.HasAttempts(2),
		perform.Log(log, slog.LevelDebug, "[ERROR] request failed, retrying"),
		perform.ExponentialBackoff(1*time.Second),
	)
	return &RetryRoundTripper{
		timeout: timeout,
		trip:    trip,
		cancels: make(map[int64]func()),
	}
}

func (rt *RetryRoundTripper) Start(ctx context.Context) {
	<-ctx.Done()
	rt.mu.Lock()
	defer rt.mu.Unlock()
	for _, cancel := range rt.cancels {
		cancel()
	}
	clear(rt.cancels)
}

func (rt *RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(context.Background())
	rt.mu.Lock()
	id := rt.cancelId
	rt.cancelId++
	rt.cancels[id] = cancel
	rt.mu.Unlock()
	defer func() {
		rt.mu.Lock()
		delete(rt.cancels, id)
		rt.mu.Unlock()
		cancel()
	}()
	return rt.trip(ctx, req)
}

func isRetryable(_ context.Context) func(resp *http.Response, err error) bool {
	return func(resp *http.Response, err error) bool {
		var netErr net.Error
		return errors.Is(err, context.DeadlineExceeded) ||
			(errors.As(err, &netErr) && netErr.Timeout()) ||
			(resp != nil && resp.StatusCode >= 500)
	}
}
