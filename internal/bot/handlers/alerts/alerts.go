package alerts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type AlertsProvider interface {
	Alerts(ctx context.Context, provider string) (ps2.Loaded[ps2.Alerts], error)
	AlertsByWorldId(ctx context.Context, provider string, worldId ps2.WorldId) (ps2.Loaded[ps2.Alerts], error)
}

func New(alertsProvider AlertsProvider) handlers.InteractionHandler {
	return handlers.DeferredResponse(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		const op = "handlers.alerts"
		log = log.With(slog.String("op", op))
		opts := i.ApplicationCommandData().Options
		log.Debug("command options", slog.Any("options", opts))
		var worldId ps2.WorldId
		var provider string
		for _, opt := range opts {
			switch opt.Name {
			case "server":
				worldId = ps2.WorldId(opt.IntValue())
			case "provider":
				provider = opt.StringValue()
			}
		}
		log.Debug("parsed options", slog.Int("world_id", int(worldId)), slog.String("provider", provider))
		if worldId > 0 {
			log.Debug("getting world alerts")
			alerts, err := alertsProvider.AlertsByWorldId(ctx, provider, worldId)
			if err != nil {
				return nil, fmt.Errorf("%s error getting alerts: %w", op, err)
			}
			worldName := ps2.WorldNameById(worldId)
			embed := []*discordgo.MessageEmbed{
				render.RenderWorldDetailedAlerts(worldName, alerts),
			}
			return &discordgo.WebhookEdit{
				Embeds: &embed,
			}, nil
		}
		log.Debug("getting global alerts")
		alerts, err := alertsProvider.Alerts(ctx, provider)
		if err != nil {
			return nil, fmt.Errorf("error getting alerts: %q", err)
		}
		embed := render.RenderAlerts(alerts)
		return &discordgo.WebhookEdit{
			Embeds: &embed,
		}, nil
	})
}
