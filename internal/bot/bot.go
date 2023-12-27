package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-feed/internal/ps2"
)

var commands = [2]*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Returns Pong!",
	},
	{
		Name:        "population",
		Description: "Returns the population.",
	},
}

type Bot struct {
	session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
}

func renderCommonPopulation(p ps2.CommonPopulation) string {
	builder := strings.Builder{}
	builder.Grow(60) // 17 characters per line
	if p.All == 0 {
		builder.WriteString("TR:   0 | 0.00%\nNC:   0 | 0.00%\nVS:   0 | 0.00%\n")
	} else {
		builder.WriteString(fmt.Sprintf("TR: %3d | %.2f%%\n", p.TR, float64(p.TR)/float64(p.All)*100))
		builder.WriteString(fmt.Sprintf("NC: %3d | %.2f%%\n", p.NC, float64(p.NC)/float64(p.All)*100))
		builder.WriteString(fmt.Sprintf("VS: %3d | %.2f%%\n", p.VS, float64(p.VS)/float64(p.All)*100))
		// builder.WriteString(fmt.Sprintf("Other: %3d | %.2f%\n", worldPopulation.Total.Other, float64(worldPopulation.Total.Other)/float64(worldPopulation.Total.All)*100))
	}
	return builder.String()
}

func renderWorldPopulation(worldPopulation ps2.WorldPopulation) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.Total.All),
		Value:  renderCommonPopulation(worldPopulation.Total),
		Inline: true,
	}
}

func renderPopulation(population ps2.Population, populationSource string, updatedAt time.Time) *discordgo.MessageEmbed {
	fields := make([]*discordgo.MessageEmbedField, 0, len(population.Worlds))
	for _, worldPopulation := range population.Worlds {
		fields = append(fields, renderWorldPopulation(worldPopulation))
	}
	return &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("Total population - %d", population.Total.All),
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %q", populationSource),
		},
		Timestamp: updatedAt.Format(time.RFC3339),
	}
}

func NewBot(discordToken string, service *ps2.Service) (*Bot, error) {
	session, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		log.Printf("Running on %d servers", len(s.State.Guilds))
	})
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
			if err != nil {
				log.Printf("Cannot respond to slash command %q: %v\n", i.ApplicationCommandData().Name, err)
			}
		},
		"population": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			population, err := service.Population(context.Background())
			if err != nil {
				log.Printf("Error getting population: %q", err)
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error getting population",
					},
				})
				if err != nil {
					log.Printf("Cannot respond to slash command %q: %v\n", i.ApplicationCommandData().Name, err)
				}
				return
			}
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{renderPopulation(population, service.PopulationSource(), service.UpdatedAt())},
				},
			})
			if err != nil {
				log.Printf("Cannot respond to slash command %q: %v\n", i.ApplicationCommandData().Name, err)
			}
		},
	}
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := handlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		} else {
			log.Printf("Unknown command %q", i.ApplicationCommandData().Name)
		}
	})
	return &Bot{
		session: session,
	}, nil
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return err
	}
	log.Println("Adding commands...")
	b.registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v)
		if err != nil {
			return fmt.Errorf("cannot create %q command: %q", v.Name, err)
		}
		b.registeredCommands[i] = cmd
	}
	return nil
}

func (b *Bot) Stop() error {
	for _, v := range b.registeredCommands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID)
		if err != nil {
			log.Printf("Cannot delete %q command: %q", v.Name, err)
		}
	}
	return b.session.Close()
}
