package events_module

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var subscriptionSettings = commands.SubscriptionSettings{
	Worlds:     []string{"all"},
	Characters: []string{"all"},
	EventNames: []string{
		string(events.PlayerLoginEventName),
		string(events.PlayerLogoutEventName),
		string(events.AchievementEarnedEventName),
		string(events.BattleRankUpEventName),
		string(events.DeathEventName),
		string(events.GainExperienceEventName),
		string(events.ItemAddedEventName),
		string(events.PlayerFacilityCaptureEventName),
		string(events.PlayerFacilityDefendEventName),
		string(events.SkillAddedEventName),
		string(events.VehicleDestroyEventName),
		string(events.FacilityControlEventName),
		string(events.MetagameEventEventName),
		string(events.ContinentLockEventName),
	},
}

func newStreamingClientService(
	log *logger.Logger,
	platform ps2_platforms.Platform,
	client *streaming.Client,
) module.Runnable {
	return module.NewRun(fmt.Sprintf("%s.streaming_client", platform), func(ctx context.Context) error {
		err := retryable.New(func(ctx context.Context) error {
			if err := client.Connect(ctx); err != nil {
				return err
			}
			return client.Subscribe(ctx, subscriptionSettings)
		})(
			ctx,
			while.ContextIsNotCancelled,
			perform.Log(log.Logger, slog.LevelError, "subscription failed, retrying"),
			perform.ExponentialBackoffWithRecover(1*time.Second),
		)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})
}
