package discord_events

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	"golang.org/x/text/language"
)

type ChannelLanguageLoader = loader.Keyed[discord.ChannelId, language.Tag]

type EventsPublisher struct {
	publisher             pubsub.Publisher[Event]
	log                   *logger.Logger
	wg                    sync.WaitGroup
	channelLanguageLoader ChannelLanguageLoader
}

func NewEventsPublisher(
	log *logger.Logger,
	pubsub pubsub.Publisher[Event],
	channelLanguageLoader ChannelLanguageLoader,
) *EventsPublisher {
	return &EventsPublisher{
		log:                   log,
		publisher:             pubsub,
		channelLanguageLoader: channelLanguageLoader,
	}
}

func (p *EventsPublisher) Start(ctx context.Context) {
	<-ctx.Done()
	p.wg.Wait()
}

func (p *EventsPublisher) PublishChannelLanguageUpdated(
	ctx context.Context,
	event storage.ChannelLanguageUpdated,
) {
	p.publish(ctx, ChannelLanguageUpdated{Event: event})
}

func (p *EventsPublisher) PublishChannelTrackerStarted(
	ctx context.Context,
	event stats_tracker.ChannelTrackerStarted,
) {
	p.wg.Add(1)
	go publishChannelEventTask(ctx, p, event.ChannelId, event)
}

func (p *EventsPublisher) PublishChannelTrackerStopped(
	ctx context.Context,
	event stats_tracker.ChannelTrackerStopped,
) {
	p.wg.Add(1)
	go publishChannelEventTask(ctx, p, event.ChannelId, event)
}

func (p *EventsPublisher) publish(ctx context.Context, event Event) {
	if err := p.publisher.Publish(event); err != nil {
		p.log.Error(ctx, "cannot publish event", slog.Any("event", event), sl.Err(err))
		return
	}
}

func publishChannelEventTask[T pubsub.EventType, E pubsub.Event[T]](
	ctx context.Context,
	p *EventsPublisher,
	channel discord.ChannelId,
	event E,
) {
	defer p.wg.Done()
	locale, err := p.channelLanguageLoader(ctx, channel)
	if err != nil {
		p.log.Error(ctx, "cannot get channel language", slog.String("channel_id", string(channel)), sl.Err(err))
		return
	}
	p.publish(ctx, localizedEvent[T, E]{
		Event:    event,
		Language: locale,
	})
}
