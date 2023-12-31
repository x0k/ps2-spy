package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/contextx"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type interactionHandler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error

func run(handler interactionHandler, s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := contextx.Await(ctx, func() error {
		err := handler(ctx, s, i)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: err.Error(),
			})
		}
		return err
	})
	if err != nil {
		log.Printf("error handling slash command %q: %q", i.ApplicationCommandData().Name, err)
	}
}

func deferredResponse(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error)) interactionHandler {
	return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			return err
		}
		data, err := handle(ctx, s, i)
		if err != nil {
			return err
		}
		_, err = s.InteractionResponseEdit(i.Interaction, data)
		return err
	}
}

func makeHandlers(service *ps2.Service) map[string]interactionHandler {
	return map[string]interactionHandler{
		"population": deferredResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
			opts := i.ApplicationCommandData().Options
			if len(opts) == 0 {
				population, err := service.Population(ctx)
				if err != nil {
					return nil, fmt.Errorf("error getting population: %q", err)
				}
				embeds := []*discordgo.MessageEmbed{
					renderPopulation(population),
				}
				return &discordgo.WebhookEdit{
					Embeds: &embeds,
				}, nil
			}
			server := opts[0].IntValue()
			population, err := service.PopulationByWorldId(ctx, ps2.WorldId(server))
			if err != nil {
				return nil, fmt.Errorf("error getting population: %q", err)
			}
			embeds := []*discordgo.MessageEmbed{
				renderWorldDetailedPopulation(population),
			}
			return &discordgo.WebhookEdit{
				Embeds: &embeds,
			}, nil
		}),
		"alerts": deferredResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
			opts := i.ApplicationCommandData().Options
			if len(opts) == 0 {
				alerts, err := service.Alerts(ctx)
				if err != nil {
					return nil, fmt.Errorf("error getting alerts: %q", err)
				}
				embeds := renderAlerts(alerts)
				return &discordgo.WebhookEdit{
					Embeds: &embeds,
				}, nil
			}
			server := opts[0].IntValue()
			alerts, err := service.AlertsByWorldId(ctx, ps2.WorldId(server))
			if err != nil {
				return nil, fmt.Errorf("error getting alerts: %q", err)
			}
			worldName := ps2.WorldNames[ps2.WorldId(server)]
			if worldName == "" {
				worldName = fmt.Sprintf("World %d", server)
			}
			embed := []*discordgo.MessageEmbed{
				renderWorldDetailedAlerts(worldName, alerts),
			}
			return &discordgo.WebhookEdit{
				Embeds: &embed,
			}, nil
		}),
	}
}
