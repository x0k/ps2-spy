package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/x0k/ps2-spy/internal/lib/loader"
)

func InstrumentQueriedLoaderWithCounterMetric[Q any, T any](
	counter *prometheus.CounterVec,
	loader loader.Queried[Q, T],
) loader.Queried[Q, T] {
	if counter == nil {
		return loader
	}
	return func(ctx context.Context, query Q) (T, error) {
		value, err := loader(ctx, query)
		labels := prometheus.Labels{
			"status": string(SuccessStatus),
		}
		if err != nil {
			labels["status"] = string(ErrorStatus)
		}
		counter.With(labels).Inc()
		return value, err
	}
}

func InstrumentQueriedLoaderWithFlightMetric[Q any, T any](
	gauge *prometheus.Gauge,
	loader loader.Queried[Q, T],
) loader.Queried[Q, T] {
	if gauge == nil {
		return loader
	}
	g := *gauge
	return func(ctx context.Context, query Q) (T, error) {
		g.Inc()
		defer g.Dec()
		return loader(ctx, query)
	}
}

func InstrumentMultiKeyedLoaderWithSubjectsCounter[K comparable, T any](
	counter *prometheus.CounterVec,
	loader loader.Multi[K, T],
) loader.Multi[K, T] {
	if counter == nil {
		return loader
	}
	return func(ctx context.Context, keys []K) (map[K]T, error) {
		res, err := loader(ctx, keys)
		counter.With(prometheus.Labels{
			"subject": string(RequestedSubject),
		}).Add(float64(len(keys)))
		counter.With(prometheus.Labels{
			"subject": string(LoadedSubject),
		}).Add(float64(len(res)))
		return res, err
	}
}
