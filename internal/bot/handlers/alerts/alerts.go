package alerts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	alertsLoader loaders.KeyedLoader[string, ps2.Alerts],
	worldAlertsLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], ps2.Alerts],
) handlers.InteractionHandler {
	return handlers.DeferredResponse(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		const op = "handlers.alerts"
		log = log.With(slog.String("op", op))
		opts := i.ApplicationCommandData().Options
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
			alerts, err := worldAlertsLoader.Load(ctx, loaders.NewMultiLoaderQuery(provider, worldId))
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
		alerts, err := alertsLoader.Load(ctx, provider)
		if err != nil {
			return nil, fmt.Errorf("error getting alerts: %q", err)
		}
		embed := render.RenderAlerts(alerts)
		return &discordgo.WebhookEdit{
			Embeds: &embed,
		}, nil
	})
}
