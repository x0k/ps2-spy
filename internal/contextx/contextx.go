package contextx

import "context"

func Go[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	await := make(chan struct{})
	var result T
	var err error
	go func() {
		defer close(await)
		result, err = fn()
	}()
	select {
	case <-ctx.Done():
		return result, ctx.Err()
	case <-await:
		return result, err
	}
}
