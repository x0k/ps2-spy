package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/honu"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var (
	discord_token string
)

func init() {
	flag.StringVar(&discord_token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	httpClient := &http.Client{}
	honuClient := honu.NewClient("https://wt.honu.pw", httpClient)
	honuClient.Start()
	defer honuClient.Stop()
	ps2Service := ps2.NewService(
		ps2.NewHonuPopulationProvider(honuClient),
		ps2.NewHonuAlertsProvider(honuClient),
	)
	ps2Service.Start()
	defer ps2Service.Stop()
	b, err := bot.NewBot(discord_token, ps2Service)
	if err != nil {
		log.Fatalln(err)
	}
	defer b.Stop()

	log.Println("Bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Gracefully shutting down.")
}
