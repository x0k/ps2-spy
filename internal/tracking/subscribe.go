package tracking

import (
	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

func Subscribe[E Event](
	postStopper module.PostStopper,
	subs pubsub.SubscriptionsManager[EventType],
) <-chan E {
	return pubsub_adapters.Subscribe[EventType, E](postStopper, subs)
}
