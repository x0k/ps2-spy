package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/x0k/ps2-feed/internal/bot"
	"github.com/x0k/ps2-feed/internal/honu"
	"github.com/x0k/ps2-feed/internal/ps2"
)

var (
	discord_token string
)

func init() {
	flag.StringVar(&discord_token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	httpClient := http.DefaultClient
	ps2Service := ps2.NewService(ps2.NewHonuPopulationProvider(honu.NewClient("https://wt.honu.pw", httpClient)))
	b, err := bot.NewBot(discord_token, ps2Service)
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
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Gracefully shutting down.")
}
