package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/x0k/ps2-feed/internal/bot"
)

var (
	discord_token string
)

func init() {
	flag.StringVar(&discord_token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	b, err := bot.NewBot(discord_token)
	if err != nil {
		log.Fatalln(err)
	}
	err = b.Start()
	if err != nil {
		log.Fatalln("error opening connection: ", err)
	}
	defer b.Stop()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop

	log.Println("Gracefully shutting down.")
}
