package main

import (
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retry"
)

func setupPs2Events(s *Setup, cfg *config.BotConfig) {
	ps2pcStream := streaming.NewClient(
		s.log,
		"wss://push.planetside2.com/streaming",
		streaming.Ps2_env,
		cfg.ServiceId,
	)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log := s.log.With(
			slog.String("component", "ps2_events"),
			slog.String("platform", streaming.Ps2_env),
		)
		retry.RetryWhileWithRecover(retry.Retryable{
			Try: func() error {
				err := ps2pcStream.Connect(s.ctx)
				if err != nil {
					log.Error("failed to connect to websocket", sl.Err(err))
					return err
				}
				defer func() {
					if err := ps2pcStream.Close(); err != nil {
						log.Error("failed to close websocket", sl.Err(err))
					}
				}()
				return ps2pcStream.Subscribe(s.ctx, ps2commands.SubscriptionSettings{
					Worlds: []string{"1", "10", "13", "17", "19", "40"},
					EventNames: []string{
						ps2events.PlayerLoginEventName,
						ps2events.PlayerLogoutEventName,
					},
				})
			},
			While: retry.ContextIsNotCanceled,
			BeforeSleep: func(d time.Duration) {
				log.Debug("retry to connect", slog.Duration("after", d))
			},
		})
	}()
}
