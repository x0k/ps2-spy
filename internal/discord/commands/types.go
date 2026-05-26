package discord_commands

import (
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type ChannelStore struct {
	Loader                       ChannelLoader
	LanguageSaver                ChannelLanguageSaver
	CharacterNotificationsSaver  ChannelCharacterNotificationsSaver
	OutfitNotificationsSaver    ChannelOutfitNotificationsSaver
	TitleUpdatesSaver            ChannelTitleUpdatesSaver
	DefaultTimezoneSaver         ChannelDefaultTimezoneSaver
}

type StatsTaskStore struct {
	Loader  ChannelStatsTrackerTasksLoader
	Creator ChannelStatsTrackerTaskCreator
	Remover ChannelStatsTrackerTaskRemover
	ById    StatsTrackerTaskLoader
	Updater ChannelStatsTrackerTaskUpdater
}

type ProviderRegistry[T any] struct {
	Loaders  map[string]T
	Priority []string
}

func NewProviderRegistry[T any](
	loaders map[string]T,
	priority []string,
) ProviderRegistry[T] {
	return ProviderRegistry[T]{
		Loaders:  loaders,
		Priority: priority,
	}
}

type PopulationProviders struct {
	ProviderRegistry[loader.Simple[meta.Loaded[ps2.WorldsPopulation]]]
}

type WorldPopulationProviders struct {
	ProviderRegistry[loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]]
}

type AlertsProviders struct {
	ProviderRegistry[loader.Simple[meta.Loaded[ps2.Alerts]]]
}
