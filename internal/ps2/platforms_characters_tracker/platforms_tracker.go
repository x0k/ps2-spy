package ps2_platforms_characters_tracker

import (
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Tracker struct {
	platforms map[ps2_platforms.Platform]*charactersTracker
}

func New(
	log *logger.Logger,
	charactersLoader CharacterLoader,
	publisher pubsub.Publisher[ps2.Event],
	mt *metrics.Metrics,
) *Tracker {
	platforms := make(map[ps2_platforms.Platform]*charactersTracker, len(ps2_platforms.Platforms))
	for _, platform := range ps2_platforms.Platforms {
		platforms[platform] = newCharactersTracker(
			log.With(slog.String("platform", string(platform))),
			platform,
			ps2.PlatformWorldIds[platform],
			charactersLoader,
			publisher,
			mt,
		)
	}
	return &Tracker{
		platforms: platforms,
	}
}

func (t *Tracker) HandlerLogin(ctx)
