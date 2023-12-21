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

func main() {
	ds, err := discordgo.New("Bot " + discord_token)
	if err != nil {
		log.Fatalln("error creating Discord session: ", err)
	}
	ds.AddHandler(handleMessageCreate)

	ds.Identify.Intents = discordgo.IntentGuildMessages

	err = ds.Open()
	if err != nil {
		log.Fatalln("error opening connection: ", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, os.Kill)
	<-sc

	ds.Close()
}

func handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "/ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			log.Println("error sending message:", err)
		}
	}
}
