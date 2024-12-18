package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
)

type RetryRoundTripper struct {
	log  *slog.Logger
	trip func(context.Context, *http.Request, ...any) (*http.Response, error)
}

func NewRetryRoundTripper(
	log *slog.Logger,
	roundTripper http.RoundTripper,
) *RetryRoundTripper {
	trip := retryable.NewWithArg(
		func(_ context.Context, req *http.Request) (*http.Response, error) {
			return roundTripper.RoundTrip(req)
		},
		isRetryable,
		while.ContextIsNotCancelled,
	)
	return &RetryRoundTripper{
		log:  log,
		trip: trip,
	}
}

func (rt *RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.trip(
		req.Context(),
		req,
		while.HasAttempts(2),
		perform.Log(rt.log, slog.LevelDebug, "[ERROR] request failed, retrying"),
		perform.ExponentialBackoff(1*time.Second),
	)
}

func isRetryable(resp *http.Response, err error) bool {
	var netErr net.Error
	return errors.Is(err, context.DeadlineExceeded) ||
		(errors.As(err, &netErr) && netErr.Timeout()) ||
		(resp != nil && resp.StatusCode >= 500)
}
