package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type Handler = pubsub.Handler[discord_events.EventType]

type handler[T discord_events.EventType, E pubsub.Event[T]] struct {
	m      *HandlersManager
	handle func(context.Context, *discordgo.Session, E) error
}

func (h *handler[T, E]) Type() T {
	var e E
	return e.Type()
}

func (h *handler[T, E]) Handle(event pubsub.Event[T]) error {
	h.m.wg.Add(1)
	go func() {
		defer h.m.wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), h.m.handlersTimeout)
		defer cancel()
		defer h.m.addCancel(cancel)()
		if err := h.handle(ctx, h.m.session, event.(E)); err != nil {
			h.m.log.Error(ctx, "cannot handle event", sl.Err(err))
		}
	}()
	return nil
}

func newHandler[T discord_events.EventType, E pubsub.Event[T]](
	m *HandlersManager,
	handle func(context.Context, *discordgo.Session, E) error,
) *handler[T, E] {
	return &handler[T, E]{
		m:      m,
		handle: handle,
	}
}
