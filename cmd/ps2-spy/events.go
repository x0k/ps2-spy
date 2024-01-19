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
	"github.com/x0k/ps2-spy/internal/lib/retry"
)

func startEventsClient(ctx context.Context, cfg *config.Config, env string, settings ps2commands.SubscriptionSettings) *streaming.Client {
	const op = "startEventsClient"
	log := infra.OpLogger(ctx, op).With(slog.String("env", env))
	client := streaming.NewClient(
		log,
		"wss://push.planetside2.com/streaming",
		env,
		cfg.CensusServiceId,
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
	return client
}

func startPs2EventsPublisher(ctx context.Context, cfg *config.Config) (*ps2events.Publisher, error) {
	const op = "startPs2EventsPublisher"
	log := infra.OpLogger(ctx, op)
	eventsPublisher := ps2events.NewPublisher()
	msgPublisher := ps2messages.NewPublisher()
	serviceMsg := make(chan ps2messages.ServiceMessage[map[string]any])
	serviceMsgUnSub, err := msgPublisher.AddHandler(serviceMsg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer serviceMsgUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-serviceMsg:
				err := eventsPublisher.Publish(msg.Payload)
				if err != nil {
					log.Error("failed to publish event", slog.Any("event", msg.Payload), sl.Err(err))
				}
			}
		}
	}()
	pcEventsClient := startEventsClient(ctx, cfg, streaming.Ps2_env, ps2commands.SubscriptionSettings{
		Worlds: []string{"1", "10", "13", "17", "19", "40"},
		EventNames: []string{
			ps2events.PlayerLoginEventName,
			ps2events.PlayerLogoutEventName,
		},
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pcEventsClient.Msg:
				err := msgPublisher.Publish(msg)
				if err != nil {
					log.Error("failed to publish message", slog.Any("message", msg), sl.Err(err))
				}
			}
		}
	}()
	return eventsPublisher, nil
}
