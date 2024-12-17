package retryable2

import (
	"context"
)

func New(
	f func(ctx context.Context) error,
	options ...any,
) func(ctx context.Context) error {
	conditions := make([]func(error) bool, 0, len(options))
	actions := make([]func(context.Context, error), 0, len(options))
	for _, option := range options {
		switch v := option.(type) {
		case func(error) bool:
			conditions = append(conditions, v)
		case func(context.Context, error):
			actions = append(actions, v)
		}
	}
	return func(ctx context.Context) error {
		for {
			err := f(ctx)
			for _, condition := range conditions {
				if !condition(err) {
					return err
				}
			}
			for _, action := range actions {
				action(ctx, err)
			}
		}
	}
}

type rError[R any] struct {
	err error
	res R
}

func (e rError[R]) Error() string {
	return e.err.Error()
}

func NewWithReturn[R any](
	f func(ctx context.Context) (R, error),
	options ...any,
) func(ctx context.Context) (R, error) {
	opts := make([]any, 0, len(options))
	for _, option := range options {
		switch v := option.(type) {
		case func(error) bool:
			opts = append(opts, func(err error) bool {
				rErr := err.(rError[R])
				return v(rErr.err)
			})
		case func(R, error) bool:
			opts = append(opts, func(err error) bool {
				rErr := err.(rError[R])
				return v(rErr.res, rErr.err)
			})
		case func(context.Context, error):
			opts = append(opts, func(ctx context.Context, err error) {
				rErr := err.(rError[R])
				v(ctx, rErr.err)
			})
		case func(context.Context, R, error):
			opts = append(opts, func(ctx context.Context, err error) {
				rErr := err.(rError[R])
				v(ctx, rErr.res, rErr.err)
			})
		}
	}
	rt := New(func(ctx context.Context) error {
		result, err := f(ctx)
		return rError[R]{err: err, res: result}
	}, opts...)
	return func(ctx context.Context) (R, error) {
		rErr := rt(ctx).(rError[R])
		return rErr.res, rErr.err
	}
}

type aContext[A any] struct {
	context.Context
	arg A
}

func NewWithArg[A any, R any](
	f func(ctx context.Context, arg A) (R, error),
	options ...any,
) func(ctx context.Context, arg A) (R, error) {
	rt := NewWithReturn(func(ctx context.Context) (R, error) {
		return f(ctx, ctx.(aContext[A]).arg)
	}, options...)
	return func(ctx context.Context, arg A) (R, error) {
		return rt(aContext[A]{Context: ctx, arg: arg})
	}
}