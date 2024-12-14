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

func (h handler[T, E]) Handle(event pubsub.Event[T]) error {
	h <- event.(E)
	return nil
}

func Subscribe[T pubsub.EventType, E pubsub.Event[T]](
	postStopper module.PostStopper,
	subs pubsub.SubscriptionsManager[T],
) <-chan E {
	channel := make(chan E)
	h := handler[T, E](channel)
	unSubscribe := subs.AddHandler(h)
	postStopper.PostStop(module.NewRun(
		fmt.Sprintf("event_handler_%v", h.Type()),
		func(_ context.Context) error {
			unSubscribe()
			close(channel)
			return nil
		},
	))
	return channel
}
