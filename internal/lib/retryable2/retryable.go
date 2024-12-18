package retryable2

import (
	"context"
)

type retryOptions struct {
	conditions []func(error) bool
	actions    []func(context.Context, error)
}

func (o retryOptions) append(options []any) retryOptions {
	additional := len(options)
	if additional == 0 {
		return o
	}
	newConditions := make([]func(error) bool, len(o.conditions), len(o.conditions)+additional)
	copy(newConditions, o.conditions)
	newActions := make([]func(context.Context, error), len(o.actions), len(o.actions)+additional)
	copy(newActions, o.actions)
	for _, option := range options {
		switch v := option.(type) {
		case func(error) bool:
			newConditions = append(newConditions, v)
		case func(context.Context, error):
			newActions = append(newActions, v)
		}
	}
	return retryOptions{
		conditions: newConditions,
		actions:    newActions,
	}
}

func New(
	f func(ctx context.Context) error,
	options ...any,
) func(ctx context.Context, options ...any) error {
	first := retryOptions{}.append(options)
	return func(ctx context.Context, options ...any) error {
		final := first.append(options)
		for {
			err := f(ctx)
			for _, condition := range final.conditions {
				if !condition(err) {
					return err
				}
			}
			for _, action := range final.actions {
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

func mapOptions[R any](options []any) []any {
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
	return opts
}

func NewWithReturn[R any](
	f func(ctx context.Context) (R, error),
	options ...any,
) func(ctx context.Context, options ...any) (R, error) {
	rt := New(func(ctx context.Context) error {
		result, err := f(ctx)
		return rError[R]{err: err, res: result}
	}, mapOptions[R](options)...)
	return func(ctx context.Context, options ...any) (R, error) {
		rErr := rt(ctx, mapOptions[R](options)...).(rError[R])
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
) func(ctx context.Context, arg A, options ...any) (R, error) {
	rt := NewWithReturn(func(ctx context.Context) (R, error) {
		return f(ctx, ctx.(aContext[A]).arg)
	}, options...)
	return func(ctx context.Context, arg A, options ...any) (R, error) {
		return rt(aContext[A]{Context: ctx, arg: arg}, options...)
	}
}
