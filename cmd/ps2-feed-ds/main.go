package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	discord_token string
)

func init() {
	flag.StringVar(&discord_token, "t", "", "Bot Token")
	flag.Parse()
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Returns Pong!",
	},
}

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		})
		if err != nil {
			log.Printf("Cannot respond to slash command %s: %v\n", i.ApplicationCommandData().Name, err)
		}
	},
}

func main() {
	ds, err := discordgo.New("Bot " + discord_token)
	if err != nil {
		log.Fatalln("error creating Discord session: ", err)
	}

	ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = ds.Open()
	if err != nil {
		log.Fatalln("error opening connection: ", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := ds.ApplicationCommandCreate(ds.State.User.ID, "", v)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop

	log.Println("Gracefully shutting down.")

	for _, v := range registeredCommands {
		err := ds.ApplicationCommandDelete(ds.State.User.ID, "", v.ID)
		if err != nil {
			log.Printf("Cannot delete '%v' command: %v\n", v.Name, err)
		}
	}

	err = ds.Close()
	if err != nil {
		log.Fatalln("error closing connection: ", err)
	}
}
