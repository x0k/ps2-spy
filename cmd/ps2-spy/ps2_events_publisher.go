package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	ps2messages "github.com/x0k/ps2-spy/internal/lib/census2/streaming/messages"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/relogin_omitter"
)

func startNewPs2EventsPublisher(
	ctx context.Context,
	logger *logger.Logger,
	mt metrics.Metrics,
	cfg *config.Config,
	platform platforms.Platform,
	settings ps2commands.SubscriptionSettings,
) (*ps2events.Publisher, error) {
	log := logger.With(
		slog.String("platform", string(platform)),
	)
	// Handle messages
	messagesPublisher := ps2messages.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.Ps2MessagesPlatformPublisher,
			platform,
			publisher.New[publisher.Event](),
		),
	)
	client := streaming.NewClient(
		log.Logger,
		cfg.CensusStreamingEndpoint,
		platforms.PlatformEnvironment(platform),
		cfg.CensusServiceId,
		messagesPublisher,
	)
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		retryable.New(
			func(ctx context.Context) error {
				err := client.Connect(ctx)
				if err != nil {
					log.Error(ctx, "failed to connect to websocket", sl.Err(err), slog.Any("settings", settings))
					return err
				}
				defer client.Close()
				return client.Subscribe(ctx, settings)
			},
		).Run(
			ctx,
			while.ContextIsNotCancelled,
			perform.RecoverSuspenseDuration(1*time.Second),
			perform.Log(log.Logger, slog.LevelError, "subscription failed, retrying"),
		)
	}()
	// Handle events

	reLoginOmitter := relogin_omitter.New(
		log,
		platform,
		mt.InstrumentPlatformPublisher(
			metrics.Ps2EventsPlatformPublisher,
			platform,
			publisher.New[publisher.Event](),
		),
		mt,
	)
	reLoginOmitter.Start(ctx)
	eventsPublisher := ps2events.NewPublisher(reLoginOmitter)
	serviceMsg := make(chan ps2messages.ServiceMessage[map[string]any])
	serviceMsgUnSub := messagesPublisher.AddServiceMessageHandler(serviceMsg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer serviceMsgUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-serviceMsg:
				if err := eventsPublisher.Publish(msg.Payload); err != nil {
					log.Error(ctx, "failed to publish event", slog.Any("event", msg.Payload), sl.Err(err))
				}
			}
		}
	}()
	return eventsPublisher, nil
}
