package while

import (
	"context"
	"errors"
)

func ErrorIsHere(err error) bool {
	return err != nil
}

func HasAttempts(attempts int) func(error) bool {
	return func(err error) bool {
		attempts--
		return attempts > 0
	}
}

func ContextIsNotCancelled(err error) bool {
	return !errors.Is(err, context.Canceled)
}
