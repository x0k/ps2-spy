package iterx

import "iter"

func Map[T, R any](seq iter.Seq[T], transform func(T) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		seq(func(value T) bool {
			return yield(transform(value))
		})
	}
}
