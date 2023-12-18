package main

import (
	"flag"
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	discord_token string
)

const API_URL = "some_url"

func init() {
	flag.StringVar(&discord_token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + discord_token)
	if err != nil {
		log.Fatalln("Error creating Discord session: ", err)
	}

}
