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

func startStreamingClient(
	ctx context.Context,
	cfg *config.Config,
	client *streaming.Client,
	settings ps2commands.SubscriptionSettings,
) {
	const op = "startStreamingClient"
	log := infra.OpLogger(ctx, op).With(slog.String("env", client.Environment()))
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
}

func startPs2EventsPublisher(
	ctx context.Context,
	cfg *config.Config,
	event chan map[string]any,
	eventsPublisher *ps2events.Publisher,
) error {
	const op = "startPs2EventsPublisher"
	log := infra.OpLogger(ctx, op)
	msgPublisher := ps2messages.NewPublisher()
	serviceMsg := make(chan ps2messages.ServiceMessage[map[string]any])
	serviceMsgUnSub, err := msgPublisher.AddHandler(serviceMsg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-event:
				err := msgPublisher.Publish(msg)
				if err != nil {
					log.Error("failed to publish message", slog.Any("message", msg), sl.Err(err))
				}
			}
		}
	}()
	return nil
}
