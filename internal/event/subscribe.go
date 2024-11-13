package event

import (
	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

func Subscribe[E Event](
	preStopper module.PreStopper,
	subs pubsub.SubscriptionsManager[Type],
) <-chan E {
	return pubsub_adapters.Subscribe[Type, E](preStopper, subs)
}
