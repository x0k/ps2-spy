package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
)

type RetryRoundTripper struct {
	log     *slog.Logger
	timeout time.Duration
	trip    func(context.Context, *http.Request, ...any) (*http.Response, error)

	mu       sync.Mutex
	cancelId int64
	cancels  map[int64]func()
}

func NewRetryRoundTripper(
	log *slog.Logger,
	timeout time.Duration,
	roundTripper http.RoundTripper,
) *RetryRoundTripper {
	trip := retryable.NewWithArg(
		func(ctx context.Context, req *http.Request) (*http.Response, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return roundTripper.RoundTrip(req.WithContext(ctx))
		},
		isRetryable,
		while.ContextIsNotCancelled,
	)
	return &RetryRoundTripper{
		log:     log,
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
	return rt.trip(
		ctx, req,
		while.HasAttempts(2),
		perform.Log(rt.log, slog.LevelDebug, "[ERROR] request failed, retrying", slog.Int64("request_id", id)),
		perform.ExponentialBackoff(1*time.Second),
	)
}

func isRetryable(resp *http.Response, err error) bool {
	var netErr net.Error
	return errors.Is(err, context.DeadlineExceeded) ||
		(errors.As(err, &netErr) && netErr.Timeout()) ||
		(resp != nil && resp.StatusCode >= 500)
}
