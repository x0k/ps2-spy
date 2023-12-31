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
	honuClient.Start()
	fisuClient.Start()
	voidWellClient.Start()
	ps2liveClient.Start()
	defer honuClient.Stop()
	defer fisuClient.Stop()
	defer voidWellClient.Stop()
	defer ps2liveClient.Stop()
	worldsLoader := ps2.WithFallback(
		ps2.WithLoaded(ps2liveClient.Endpoint(), ps2.NewPS2LiveWorldsPopulationLoader(ps2liveClient)),
		ps2.WithLoaded(honuClient.Endpoint(), ps2.NewHonuWorldsPopulationLoader(honuClient)),
		ps2.WithLoaded(fisuClient.Endpoint(), ps2.NewFisuWorldsPopulationLoader(fisuClient)),
		ps2.WithLoaded(voidWellClient.Endpoint(), ps2.NewVoidWellWorldsPopulationLoader(voidWellClient)),
	)
	worldLoader := ps2.WithKeyedFallback(
		ps2.WithKeyedLoaded(honuClient.Endpoint(), ps2.NewHonuWorldPopulationLoader(honuClient)),
		ps2.WithKeyedLoaded(voidWellClient.Endpoint(), ps2.NewVoidWellWorldPopulationLoader(voidWellClient)),
	)
	alertsLoader := ps2.WithFallback(
		ps2.WithLoaded(honuClient.Endpoint(), ps2.NewHonuAlertsLoader(honuClient)),
		ps2.WithLoaded(voidWellClient.Endpoint(), ps2.NewVoidWellAlertsLoader(voidWellClient)),
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
