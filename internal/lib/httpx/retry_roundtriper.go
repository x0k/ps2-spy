package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/retryable2"
	"github.com/x0k/ps2-spy/internal/lib/retryable2/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable2/while"
)

type RetryRoundTripper func(context.Context, *http.Request) (*http.Response, error)

func (rrt RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rrt(req.Context(), req)
}

func NewRetryRoundTripper(
	log *slog.Logger,
	roundTripper http.RoundTripper,
) RetryRoundTripper {
	rt := retryable2.NewWithArg(
		func(
			_ context.Context,
			req *http.Request,
		) (*http.Response, error) {
			return roundTripper.RoundTrip(req)
		},
		isRetryable,
		while.ContextIsNotCancelled,
		while.HasAttempts(2),
		perform.Log(log, slog.LevelDebug, "[ERROR] request failed, retrying"),
		perform.ExponentialBackoff(1*time.Second),
	)
	return RetryRoundTripper(rt)
}

func isRetryable(resp *http.Response, err error) bool {
	var netErr net.Error
	return (errors.As(err, &netErr) && netErr.Timeout()) ||
		(resp != nil && resp.StatusCode >= 500)
}
