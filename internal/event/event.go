package event

import "github.com/x0k/ps2-spy/internal/lib/pubsub"

type Type string

type Event pubsub.Event[Type]

type Publisher pubsub.Publisher[Type]
