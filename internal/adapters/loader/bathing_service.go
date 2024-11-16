package loader_adapters

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/module"
)

func NewBatchingService[K comparable, T any](
	name string,
	batched *loader.Batched[K, T],
) module.Service {
	return module.NewService(name, func(ctx context.Context) error {
		batched.Start(ctx)
		return nil
	})
}
