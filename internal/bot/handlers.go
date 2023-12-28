package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-feed/internal/contextx"
	"github.com/x0k/ps2-feed/internal/ps2"
)

type interactionHandler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error

func run(handler interactionHandler, s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := contextx.Await(ctx, func() error {
		err := handler(ctx, s, i)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
				},
			})
		}
		return err
	})
	if err != nil {
		log.Printf("Error handling slash command %q: %q", i.ApplicationCommandData().Name, err)
	}
}

func simpleResponse(handle func(ctx context.Context) (*discordgo.InteractionResponseData, error)) interactionHandler {
	return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		data, err := handle(ctx)
		if err != nil {
			return err
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		})
		if err != nil {
			log.Printf("Cannot respond to slash command %q: %v", i.ApplicationCommandData().Name, err)
		}
		return nil
	}
}

func makeHandlers(service *ps2.Service) map[string]interactionHandler {
	return map[string]interactionHandler{
		"ping": simpleResponse(func(ctx context.Context) (*discordgo.InteractionResponseData, error) {
			return &discordgo.InteractionResponseData{Content: "Pong!"}, nil
		}),
		"population": simpleResponse(func(ctx context.Context) (*discordgo.InteractionResponseData, error) {
			population, err := service.Population(ctx)
			if err != nil {
				return nil, fmt.Errorf("error getting population: %q", err)
			}
			return &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					renderPopulation(population, service.PopulationSource(), service.UpdatedAt()),
				},
			}, nil
		}),
	}
}
