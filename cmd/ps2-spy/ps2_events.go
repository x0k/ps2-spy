package main

import (
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

func startEventsClient(s *Setup, cfg *config.BotConfig, env string, settings ps2commands.SubscriptionSettings) *streaming.Client {
	client := streaming.NewClient(
		s.log,
		"wss://push.planetside2.com/streaming",
		env,
		cfg.ServiceId,
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
				defer func() {
					if err := client.Close(); err != nil {
						log.Error("failed to close websocket", sl.Err(err))
					}
				}()
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

func startPrinter(s *Setup, cfg *config.BotConfig) {
	eventPublisher := ps2events.NewPublisher(s.log)
	logins := make(chan ps2events.PlayerLogin)
	loginsUnSub, err := eventPublisher.AddHandler(logins)
	if err != nil {
		s.log.Error("failed to add login handler", sl.Err(err))
		return
	}
	logouts := make(chan ps2events.PlayerLogout)
	logoutsUnSub, err := eventPublisher.AddHandler(logouts)
	if err != nil {
		s.log.Error("failed to add logout handler", sl.Err(err))
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer loginsUnSub()
		defer logoutsUnSub()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg := <-logins:
				s.log.Debug("login msg", slog.Any("msg", msg))
			case msg := <-logouts:
				s.log.Debug("logout msg", slog.Any("msg", msg))
			}
		}
	}()

	msgPublisher := ps2messages.NewPublisher(s.log)
	srv := make(chan ps2messages.ServiceMessage[map[string]any])
	srvUnSub, err := msgPublisher.AddHandler(srv)
	if err != nil {
		s.log.Error("failed to add message handler", sl.Err(err))
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer srvUnSub()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg := <-srv:
				eventPublisher.Publish(msg.Payload)
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

}
