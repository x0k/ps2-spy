package discord_commands

import (
	"maps"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	log *logger.Logger,
	messages discord.LocalizedMessages,
	populationLoaders map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]],
	worldPopulationLoaders map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]],
) []*discord.Command {
	return []*discord.Command{
		NewAbout(messages),
		NewPopulation(
			log.With(sl.Component("population_command")),
			messages,
			maps.Keys(populationLoaders),
			maps.Keys(worldPopulationLoaders),
			nil,
			nil,
		),
	}
}
