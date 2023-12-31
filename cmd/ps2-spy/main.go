package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/fisu"
	"github.com/x0k/ps2-spy/internal/honu"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/loaders"
	"github.com/x0k/ps2-spy/internal/ps2alerts"
	"github.com/x0k/ps2-spy/internal/ps2live"
	"github.com/x0k/ps2-spy/internal/voidwell"
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
	fisuClient := fisu.NewClient("https://ps2.fisu.pw", httpClient)
	voidWellClient := voidwell.NewClient("https://api.voidwell.com", httpClient)
	ps2liveClient := ps2live.NewPopulationClient("https://agg.ps2.live", httpClient)
	ps2alertsClient := ps2alerts.NewClient("https://api.ps2alerts.com/", httpClient)
	honuClient.Start()
	fisuClient.Start()
	voidWellClient.Start()
	ps2liveClient.Start()
	ps2alertsClient.Start()
	defer honuClient.Stop()
	defer fisuClient.Stop()
	defer voidWellClient.Stop()
	defer ps2liveClient.Stop()
	defer ps2alertsClient.Stop()
	worldsLoader := ps2.WithFallback(
		"Worlds",
		ps2.WithLoaded(loaders.NewPS2LiveWorldsPopulationLoader(ps2liveClient)),
		ps2.WithLoaded(loaders.NewHonuWorldsPopulationLoader(honuClient)),
		ps2.WithLoaded(loaders.NewFisuWorldsPopulationLoader(fisuClient)),
		ps2.WithLoaded(loaders.NewVoidWellWorldsPopulationLoader(voidWellClient)),
	)
	worldLoader := ps2.WithKeyedFallback(
		"World",
		ps2.WithKeyedLoaded(loaders.NewHonuWorldPopulationLoader(honuClient)),
		ps2.WithKeyedLoaded(loaders.NewVoidWellWorldPopulationLoader(voidWellClient)),
	)
	alertsLoader := ps2.WithFallback(
		"Alerts",
		ps2.WithLoaded(loaders.NewPS2AlertsAlertsLoader(ps2alertsClient)),
		ps2.WithLoaded(loaders.NewHonuAlertsLoader(honuClient)),
		ps2.WithLoaded(loaders.NewVoidWellAlertsLoader(voidWellClient)),
	)
	worldsLoader.Start()
	worldLoader.Start()
	alertsLoader.Start()
	defer worldsLoader.Stop()
	defer worldLoader.Stop()
	defer alertsLoader.Stop()
	ps2Service := ps2.NewService(
		worldsLoader,
		worldLoader,
		alertsLoader,
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
