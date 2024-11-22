package census_data_provider

import (
	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

func Subscribe[E events.Event](
	postStopper module.PostStopper,
	subs pubsub.SubscriptionsManager[events.EventType],
) <-chan E {
	return pubsub_adapters.Subscribe[events.EventType, E](postStopper, subs)
}
