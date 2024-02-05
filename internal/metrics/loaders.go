package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
)

type instrumentedQueriedLoaderByCounter[Q any, T any] struct {
	loaders.QueriedLoader[Q, T]
	counter *prometheus.CounterVec
}

func (l *instrumentedQueriedLoaderByCounter[Q, T]) Load(ctx context.Context, query Q) (T, error) {
	value, err := l.QueriedLoader.Load(ctx, query)
	labels := prometheus.Labels{
		"status": string(SuccessStatus),
	}
	if err != nil {
		labels["status"] = string(ErrorStatus)
	}
	l.counter.With(labels).Inc()
	return value, err
}

func InstrumentQueriedLoaderWithCounterMetric[Q any, T any](
	counter *prometheus.CounterVec,
	loader loaders.QueriedLoader[Q, T],
) loaders.QueriedLoader[Q, T] {
	if counter == nil {
		return loader
	}
	return &instrumentedQueriedLoaderByCounter[Q, T]{
		QueriedLoader: loader,
		counter:       counter,
	}
}

type instrumentedQueriedLoaderByInFlightCounter[Q any, T any] struct {
	loaders.QueriedLoader[Q, T]
	gauge prometheus.Gauge
}

func (l *instrumentedQueriedLoaderByInFlightCounter[Q, T]) Load(ctx context.Context, query Q) (T, error) {
	l.gauge.Inc()
	defer l.gauge.Dec()
	return l.QueriedLoader.Load(ctx, query)
}

func InstrumentQueriedLoaderWithFlightMetric[Q any, T any](
	gauge *prometheus.Gauge,
	loader loaders.QueriedLoader[Q, T],
) loaders.QueriedLoader[Q, T] {
	if gauge == nil {
		return loader
	}
	return &instrumentedQueriedLoaderByInFlightCounter[Q, T]{
		QueriedLoader: loader,
		gauge:         *gauge,
	}
}

type instrumentedMultiKeyedLoaderBySubjectsCounter[K comparable, T any] struct {
	loaders.QueriedLoader[[]K, map[K]T]
	counter *prometheus.CounterVec
}

func (l *instrumentedMultiKeyedLoaderBySubjectsCounter[K, T]) Load(ctx context.Context, keys []K) (map[K]T, error) {
	res, err := l.QueriedLoader.Load(ctx, keys)
	l.counter.With(prometheus.Labels{
		"subject": string(RequestedSubject),
	}).Add(float64(len(keys)))
	l.counter.With(prometheus.Labels{
		"subject": string(LoadedSubject),
	}).Add(float64(len(res)))
	return res, err
}

func InstrumentMultiKeyedLoaderWithSubjectsCounter[K comparable, T any](
	counter *prometheus.CounterVec,
	loader loaders.QueriedLoader[[]K, map[K]T],
) loaders.QueriedLoader[[]K, map[K]T] {
	if counter == nil {
		return loader
	}
	return &instrumentedMultiKeyedLoaderBySubjectsCounter[K, T]{
		QueriedLoader: loader,
		counter:       counter,
	}
}
