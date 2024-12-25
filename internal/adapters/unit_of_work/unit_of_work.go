package unit_of_work_adapters

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/unit_of_work"
)

type UnitOfWork[T any, E any] interface {
	unit_of_work.UnitOfWork[T]
	pubsub.Publisher[E]
}

type Factory[T any, E any] func(context.Context) (UnitOfWork[T, E], error)
