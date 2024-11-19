package discord_commands

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type commands struct {
	name                  string
	commands              []*discord.Command
	populationLoader      *populationLoader
	worldPopulationLoader *worldPopulationLoader
}

func New(
	name string,
	log *logger.Logger,
	messages discord.LocalizedMessages,
	populationLoaders map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]],
	populationLoadersPriority []string,
	worldPopulationLoaders map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]],
	worldPopulationLoadersPriority []string,
) *commands {
	populationLoader := newPopulationLoader(
		fmt.Sprintf("%s.population_loader", name),
		log.With(sl.Component("population_loader")),
		populationLoaders,
		populationLoadersPriority,
	)
	worldPopulationLoader := newWorldPopulationLoader(
		fmt.Sprintf("%s.world_population_loader", name),
		log.With(sl.Component("world_population_loader")),
		worldPopulationLoaders,
		worldPopulationLoadersPriority,
	)
	return &commands{
		name:                  name,
		populationLoader:      populationLoader,
		worldPopulationLoader: worldPopulationLoader,
		commands: []*discord.Command{
			NewAbout(messages),
			NewPopulation(
				log.With(sl.Component("population_command")),
				messages,
				populationLoader.load,
				slices.Values(populationLoadersPriority),
				worldPopulationLoader.load,
				slices.Values(worldPopulationLoadersPriority),
			),
		},
	}
}

func (c *commands) Name() string {
	return c.name
}

func (c *commands) Commands() []*discord.Command {
	return c.commands
}

func (c *commands) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.worldPopulationLoader.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		c.populationLoader.Start(ctx)
	}()
	<-ctx.Done()
	wg.Wait()
	return nil
}
