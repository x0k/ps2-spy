package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	ps2messages "github.com/x0k/ps2-spy/internal/lib/census2/streaming/messages"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retry"
)

func startEventsClient(s *setup, cfg *config.Config, env string, settings ps2commands.SubscriptionSettings) *streaming.Client {
	client := streaming.NewClient(
		s.log,
		"wss://push.planetside2.com/streaming",
		env,
		cfg.CensusServiceId,
	)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log := s.log.With(
			slog.String("component", "events_client_starter"),
			slog.String("platform", env),
		)
		retry.RetryWhileWithRecover(retry.Retryable{
			Try: func() error {
				err := client.Connect(s.ctx)
				if err != nil {
					log.Error("failed to connect to websocket", sl.Err(err))
					return err
				}
				defer client.Close()
				return client.Subscribe(s.ctx, settings)
			},
			While: retry.ContextIsNotCanceled,
			BeforeSleep: func(d time.Duration) {
				log.Debug("retry to connect", slog.Duration("after", d))
			},
		})
	}()
	return client
}

func startPs2EventsPublisher(s *setup, cfg *config.Config) (*ps2events.Publisher, error) {
	const op = "startPs2EventsPublisher"
	eventsPublisher := ps2events.NewPublisher(s.log)
	msgPublisher := ps2messages.NewPublisher(s.log)
	serviceMsg := make(chan ps2messages.ServiceMessage[map[string]any])
	serviceMsgUnSub, err := msgPublisher.AddHandler(serviceMsg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer serviceMsgUnSub()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg := <-serviceMsg:
				eventsPublisher.Publish(msg.Payload)
			}
		}
	}()
	pc := startEventsClient(s, cfg, streaming.Ps2_env, ps2commands.SubscriptionSettings{
		Worlds: []string{"1", "10", "13", "17", "19", "40"},
		EventNames: []string{
			ps2events.PlayerLoginEventName,
			ps2events.PlayerLogoutEventName,
		},
	})
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg := <-pc.Msg:
				msgPublisher.Publish(msg)
			}
		}
	}()
	return eventsPublisher, nil
}
