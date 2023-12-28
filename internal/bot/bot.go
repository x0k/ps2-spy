package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-feed/internal/ps2"
)

func serverNames() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(ps2.WorldNames))
	for k, v := range ps2.WorldNames {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: int(k),
		})
	}
	return choices
}

var commands = [2]*discordgo.ApplicationCommand{
	{
		Name:        "population",
		Description: "Returns the population.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "server",
				Description: "Server name",
				Required:    false,
				Choices:     serverNames(),
			},
		},
	},
	{
		Name:        "alerts",
		Description: "Returns the alerts.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "server",
				Description: "Server name",
				Required:    false,
				Choices:     serverNames(),
			},
		},
	},
}

type Bot struct {
	session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
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
	handlers := makeHandlers(service)
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := handlers[i.ApplicationCommandData().Name]; ok {
			go run(handler, s, i)
		} else {
			log.Printf("Unknown command %q", i.ApplicationCommandData().Name)
		}
	})
	err = session.Open()
	if err != nil {
		return nil, err
	}
	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for _, v := range commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", v)
		if err != nil {
			log.Printf("cannot create %q command: %q", v.Name, err)
		} else {
			registeredCommands = append(registeredCommands, cmd)
		}
	}
	return &Bot{
		session:            session,
		registeredCommands: registeredCommands,
	}, nil
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
