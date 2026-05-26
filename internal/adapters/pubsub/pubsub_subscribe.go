package pubsub_adapters

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type handler[T pubsub.EventType, E pubsub.Event[T]] chan<- E

func (h handler[T, E]) Type() T {
	var e E
	return e.Type()
}

func (h handler[T, E]) Handle(event pubsub.Event[T]) {
	h <- event.(E)
}

func SubscribeTo[T pubsub.EventType, E pubsub.Event[T]](
	postStopper module.Stopper,
	subs pubsub.SubscriptionsManager[T],
) <-chan E {
	channel := make(chan E)
	h := handler[T, E](channel)
	unSubscribe := subs.AddHandler(h)
	postStopper.OnStop(module.NewRun(
		fmt.Sprintf("event_handler_%v", h.Type()),
		func(_ context.Context) error {
			unSubscribe()
			close(channel)
			return nil
		},
	))
	return channel
}

func Listen[T pubsub.EventType, E pubsub.Event[T]](
	stopper module.Stopper,
	subs pubsub.SubscriptionsManager[T],
	name string,
	handler func(context.Context, E),
) module.Runnable {
	ch := SubscribeTo[T, E](stopper, subs)
	return module.NewRun(name, func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case e := <-ch:
				handler(ctx, e)
			}
		}
	})
}
