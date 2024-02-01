package alerts_command_handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	log *logger.Logger,
	alertsLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.Alerts]],
	worldAlertsLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.Alerts]],
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, *handlers.Error) {
		const op = "bot.handlers.command.alerts_command_handler"
		opts := i.ApplicationCommandData().Options
		var worldId ps2.WorldId
		var provider string
		for _, opt := range opts {
			switch opt.Name {
			case "server":
				worldId = ps2.WorldId(opt.StringValue())
			case "provider":
				provider = opt.StringValue()
			}
		}
		log.Debug(ctx, "parsed options", slog.String("world_id", string(worldId)), slog.String("provider", provider))
		if worldId != "" {
			log.Debug(ctx, "getting world alerts")
			alerts, err := worldAlertsLoader.Load(ctx, loaders.NewMultiLoaderQuery(provider, worldId))
			if err != nil {
				return nil, &handlers.Error{
					Msg: "Failed to get world alerts",
					Err: fmt.Errorf("%s getting world alerts: %w", op, err),
				}
			}
			worldName := ps2.WorldNameById(worldId)
			embed := []*discordgo.MessageEmbed{
				render.RenderWorldDetailedAlerts(worldName, alerts),
			}
			return &discordgo.WebhookEdit{
				Embeds: &embed,
			}, nil
		}
		log.Debug(ctx, "getting global alerts")
		alerts, err := alertsLoader.Load(ctx, provider)
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to get global alerts",
				Err: fmt.Errorf("%s getting alerts: %w", op, err),
			}
		}
		embed := render.RenderAlerts(alerts)
		return &discordgo.WebhookEdit{
			Embeds: &embed,
		}, nil
	})
}
