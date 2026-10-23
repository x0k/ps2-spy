package discord_events

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/taskx"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type ChannelLoader = loader.Keyed[discord.ChannelId, discord.Channel]

type EventsPublisher struct {
	publisher     pubsub.Publisher[Event]
	log           *logger.Logger
	tasks         *taskx.Group
	channelLoader ChannelLoader
}

func NewEventsPublisher(
	log *logger.Logger,
	pubsub pubsub.Publisher[Event],
	channelLanguageLoader ChannelLoader,
) *EventsPublisher {
	return &EventsPublisher{
		log:           log,
		tasks:         taskx.NewGroup(),
		publisher:     pubsub,
		channelLoader: channelLanguageLoader,
	}
}

func (p *EventsPublisher) Start(ctx context.Context) {
	<-ctx.Done()
	p.tasks.Close()
}

func (p *EventsPublisher) PublishChannelLanguageUpdated(
	ctx context.Context,
	event storage.ChannelLanguageSaved,
) {
	p.tasks.Go(func() {
		publishChannelEventTask(ctx, p, event.ChannelId, event)
	})
}

func (p *EventsPublisher) PublishChannelTitleUpdates(
	ctx context.Context,
	event storage.ChannelTitleUpdatesSaved,
) {
	p.tasks.Go(func() {
		publishChannelEventTask(ctx, p, event.ChannelId, event)
	})
}

func (p *EventsPublisher) PublishChannelTrackerStarted(
	ctx context.Context,
	event stats_tracker.ChannelTrackerStarted,
) {
	p.tasks.Go(func() {
		publishChannelEventTask(ctx, p, event.ChannelId, event)
	})
}

func (p *EventsPublisher) PublishChannelTrackerStopped(
	ctx context.Context,
	event stats_tracker.ChannelTrackerStopped,
) {
	p.tasks.Go(func() {
		publishChannelEventTask(ctx, p, event.ChannelId, event)
	})
}

func (p *EventsPublisher) PublishChannelTrackingSettingsUpdated(
	ctx context.Context,
	event tracking.TrackingSettingsUpdated,
) {
	p.tasks.Go(func() {
		publishChannelEventTask(ctx, p, event.ChannelId, event)
	})
}

func publishChannelEventTask[T pubsub.EventType, E pubsub.Event[T]](
	ctx context.Context,
	p *EventsPublisher,
	channelId discord.ChannelId,
	event E,
) {
	channel, err := p.channelLoader(ctx, channelId)
	if err != nil {
		p.log.Error(ctx, "cannot get channel language", slog.String("channel_id", string(channelId)), sl.Err(err))
		return
	}
	p.publisher.Publish(channelEvent[T, E]{
		Event:   event,
		Channel: channel,
	})
}
