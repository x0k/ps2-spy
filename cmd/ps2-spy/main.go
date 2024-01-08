package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/loaders/alerts"
	"github.com/x0k/ps2-spy/internal/ps2/loaders/world"
	"github.com/x0k/ps2-spy/internal/ps2/loaders/worlds"
)

var (
	config_path string
)

func init() {
	flag.StringVar(&config_path, "config", os.Getenv("CONFIG_PATH"), "Config path")
	flag.Parse()
}

func main() {
	cfg := config.MustLoad(config_path)

	httpClient := &http.Client{}
	honuClient := honu.NewClient("https://wt.honu.pw", httpClient)
	fisuClient := fisu.NewClient("https://ps2.fisu.pw", httpClient)
	voidWellClient := voidwell.NewClient("https://api.voidwell.com", httpClient)
	populationClient := population.NewClient("https://agg.ps2.live", httpClient)
	saerroClient := saerro.NewClient("https://saerro.ps2.live", httpClient)
	ps2alertsClient := ps2alerts.NewClient("https://api.ps2alerts.com", httpClient)
	censusClient := census2.NewClient("https://census.daybreakgames.com", "", httpClient)
	sanctuaryClient := census2.NewClient("https://census.lithafalcon.cc", "", httpClient)
	honuClient.Start()
	fisuClient.Start()
	voidWellClient.Start()
	populationClient.Start()
	saerroClient.Start()
	ps2alertsClient.Start()
	defer honuClient.Stop()
	defer fisuClient.Stop()
	defer voidWellClient.Stop()
	defer populationClient.Stop()
	defer saerroClient.Stop()
	defer ps2alertsClient.Stop()
	ps2Service := ps2.NewService(
		map[string]ps2.Loader[ps2.WorldsPopulation]{
			"honu":      worlds.NewHonuLoader(honuClient),
			"ps2live":   worlds.NewPS2LiveLoader(populationClient),
			"saerro":    worlds.NewSaerroLoader(saerroClient),
			"fisu":      worlds.NewFisuLoader(fisuClient),
			"sanctuary": worlds.NewSanctuaryLoader(sanctuaryClient),
			"voidwell":  worlds.NewVoidWellLoader(voidWellClient),
		},
		[]string{"honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},

		map[string]ps2.KeyedLoader[ps2.WorldId, ps2.DetailedWorldPopulation]{
			"honu":     world.NewHonuLoader(honuClient),
			"saerro":   world.NewSaerroLoader(saerroClient),
			"voidwell": world.NewVoidWellLoader(voidWellClient),
		},
		[]string{"honu", "saerro", "voidwell"},

		map[string]ps2.Loader[ps2.Alerts]{
			"ps2alerts": alerts.NewPS2AlertsLoader(ps2alertsClient),
			"honu":      alerts.NewHonuLoader(honuClient),
			"census":    alerts.NewCensusLoader(censusClient),
			"voidwell":  alerts.NewVoidWellLoader(voidWellClient),
		},
		[]string{"ps2alerts", "honu", "census", "voidwell"},
	)
	ps2Service.Start()
	defer ps2Service.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps2events := streaming.NewClient("wss://push.planetside2.com/streaming", streaming.Ps2_env, "example")
	err := ps2events.Connect(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer ps2events.Close()

	b, err := bot.NewBot(ctx, cfg.Discord.Token, ps2Service)
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
