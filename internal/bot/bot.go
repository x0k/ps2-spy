package bot

import (
	"fmt"
	"log"

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
