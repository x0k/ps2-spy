package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	ps2messages "github.com/x0k/ps2-spy/internal/lib/census2/streaming/messages"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/lib/retry"
	"github.com/x0k/ps2-spy/internal/relogin_omitter"
)

func startNewPs2EventsPublisher(
	ctx context.Context,
	cfg *config.Config,
	env string,
	settings ps2commands.SubscriptionSettings,
) (*publisher.Publisher, error) {
	const op = "startNewPs2EventsPublisher"
	log := infra.OpLogger(ctx, op)
	// Handle messages
	messagesPublisher := publisher.New(ps2messages.CastHandler)
	rawMessagesPublisher := ps2messages.NewPublisher(messagesPublisher)
	client := streaming.NewClient(
		log,
		cfg.CensusStreamingEndpoint,
		env,
		cfg.CensusServiceId,
		rawMessagesPublisher,
	)
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		retry.RetryWhileWithRecover(retry.Retryable{
			Try: func() error {
				err := client.Connect(ctx)
				if err != nil {
					log.Error("failed to connect to websocket", sl.Err(err))
					return err
				}
				defer client.Close()
				return client.Subscribe(ctx, settings)
			},
			While: retry.ContextIsNotCanceled,
			BeforeSleep: func(d time.Duration) {
				log.Debug("retry to connect", slog.Duration("after", d))
			},
		})
	}()
	// Handle events
	eventsPublisher := publisher.New(ps2events.CastHandler)
	reLoginOmitter := relogin_omitter.New(eventsPublisher)
	reLoginOmitter.Start(ctx)
	rawEventsPublisher := ps2events.NewPublisher(reLoginOmitter)
	serviceMsg := make(chan ps2messages.ServiceMessage[map[string]any])
	serviceMsgUnSub, err := messagesPublisher.AddHandler(serviceMsg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer serviceMsgUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-serviceMsg:
				if err := rawEventsPublisher.Publish(msg.Payload); err != nil {
					log.Error("failed to publish event", slog.Any("event", msg.Payload), sl.Err(err))
				}
			}
		}
	}()
	return eventsPublisher, nil
}
