package contextx

import "context"

func Await(ctx context.Context, fn func() error) error {
	await := make(chan struct{})
	var err error
	go func() {
		defer close(await)
		err = fn()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-await:
		return err
	}
}

func AwaitValue[T any](ctx context.Context, fn func() (T, error)) (T, error) {
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

func AwaitValueWithContext[T any](ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	await := make(chan struct{})
	var result T
	var err error
	go func() {
		defer close(await)
		result, err = fn(ctx)
	}()
	select {
	case <-ctx.Done():
		return result, ctx.Err()
	case <-await:
		return result, err
	}
}
