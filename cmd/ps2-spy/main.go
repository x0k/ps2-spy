package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"nhooyr.io/websocket"
)

var (
	config_path string
)

func init() {
	flag.StringVar(&config_path, "config", os.Getenv("CONFIG_PATH"), "Config path")
	flag.Parse()
}

func main() {
	cfg := config.MustLoad(config_path)
	log := mustSetupLogger(&cfg.Logger)
	log.Info("starting...", slog.String("log_level", cfg.Logger.Level))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := &Setup{
		log: log,
		ctx: ctx,
		wg:  &sync.WaitGroup{},
	}

	storage := mustSetupStorage(s, &cfg.Storage)
	_ = storage

	ps2Service := setupPs2Service(s, &cfg.Ps2Service)

	ps2pcStream := streaming.NewClient(log, "wss://push.planetside2.com/streaming", streaming.Ps2_env, "example")
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := ps2pcStream.Connect(ctx)
				if err != nil {
					log.Error("failed to connect to websocket", sl.Err(err))
					continue
				}
				defer ps2pcStream.Close()
				shouldRetry := true
				for shouldRetry {
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Second):
						err = ps2pcStream.Subscribe(ctx, ps2commands.SubscriptionSettings{
							Worlds: []string{"1", "10", "13", "17", "19", "40"},
							EventNames: []string{
								ps2events.PlayerLoginEventName,
								ps2events.PlayerLogoutEventName,
							},
						})
						if err != nil {
							log.Error("failed to subscribe to websocket", sl.Err(err))
							shouldRetry = int(websocket.CloseStatus(err)) == -1
						}
					}
				}
			}
		}
	}()

	b := mustSetupBot(s, &cfg.Bot, ps2Service)
	_ = b

	log.Info("bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("gracefully shutting down.")
	cancel()
	s.wg.Wait()
}
