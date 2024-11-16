package ps2_module

import (
	"log/slog"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/metrics"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Config struct {
	HttpClient        *http.Client
	Platform          ps2_platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
}

func New(log *logger.Logger, mt *metrics.Metrics, cfg *Config) (*module.Module, error) {
	log = log.With(slog.String("module", "ps2"))
	m := module.New(log.Logger, "ps2")

	// populationLoader := characters_tracker_population_loader.NewCharactersTrackerLoader(
	// 	log,
	// 	cfg.BotName,
	// )

	return m, nil
}
