package while

import (
	"context"
	"errors"
)

func ErrorIsHere(_ context.Context) func(err error) bool {
	return func(err error) bool {
		return err != nil
	}
}

func HasAttempts(attempts int) func(context.Context) func(error) bool {
	return func(_ context.Context) func(error) bool {
		a := attempts
		return func(err error) bool {
			a--
			return a > 0
		}
	}
}

func ContextIsNotCancelled(_ context.Context) func(err error) bool {
	return func(err error) bool {
		return !errors.Is(err, context.Canceled)
	}
}
