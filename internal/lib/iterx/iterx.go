package iterx

import "iter"

func Map[T, R any](seq iter.Seq[T], transform func(T) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		seq(func(value T) bool {
			return yield(transform(value))
		})
	}
}

// borrowed from https://pkg.go.dev/github.com/jub0bs/iterutil#Concat
func Concat[E any](seqs ...iter.Seq[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, seq := range seqs {
			for e := range seq {
				if !yield(e) {
					return
				}
			}
		}
	}
}
