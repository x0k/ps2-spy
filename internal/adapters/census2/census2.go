package census2_adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
)

type request struct {
	client     *census2.Client
	collection string
	url        string
}

var retryableExecutePrepared = retryable.NewWithArg(func(
	ctx context.Context, r request,
) (json.RawMessage, error) {
	return r.client.ExecutePrepared(ctx, r.collection, r.url)
})

func RetryableExecutePreparedAndDecode[T any](
	ctx context.Context, log *logger.Logger, c *census2.Client, collection, url string,
) ([]T, error) {
	data, err := retryableExecutePrepared(
		ctx,
		request{
			client:     c,
			collection: collection,
			url:        url,
		},
		while.HasAttempts(2),
		while.ErrorIsHere,
		while.ContextIsNotCancelled,
		perform.Log(
			log.Logger,
			slog.LevelDebug,
			"[ERROR] failed to execute prepared, retrying",
			slog.String("url", url),
		),
		perform.ExponentialBackoff(1*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute prepared: %w", err)
	}
	return census2.DecodeCollection[T](data)
}
