package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/cache/facility_cache"
	"github.com/x0k/ps2-spy/internal/cache/outfits_cache"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/loaders/alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/platform_character_names_loader"
	"github.com/x0k/ps2-spy/internal/loaders/platform_outfit_tags_loader"
	"github.com/x0k/ps2-spy/internal/loaders/platform_outfits_loader"
	"github.com/x0k/ps2-spy/internal/loaders/population_loader"
	"github.com/x0k/ps2-spy/internal/loaders/subscription_settings_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_online_entities_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_population_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_territory_control_loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type BotConfig struct {
	DiscordToken           string
	RemoveCommands         bool
	CommandHandlerTimeout  time.Duration
	Ps2EventHandlerTimeout time.Duration
	Commands               []*discordgo.ApplicationCommand
	CommandHandlers        map[string]handlers.InteractionHandler
	SubmitHandlers         map[string]handlers.InteractionHandler
}

func NewConfig(
	log *logger.Logger,
	cfg *config.Config,
	sqlStorage *sqlite.Storage,
	censusClient *census2.Client,
	facilityCache *facility_cache.StorageCache,
	popLoader *population_loader.MultiLoader,
	worldPopLoader *world_population_loader.MultiLoader,
	alertsLoader *alerts_loader.MultiLoader,
	worldAlertsLoader *world_alerts_loader.MultiLoader,
	platformWorldsTrackers map[platforms.Platform]*worlds_tracker.WorldsTracker,
	platformCharactersTrackers map[platforms.Platform]*characters_tracker.CharactersTracker,
) *BotConfig {
	subSettingsLoader := subscription_settings_loader.New(sqlStorage)
	return &BotConfig{
		DiscordToken:           cfg.DiscordToken,
		RemoveCommands:         cfg.RemoveCommands,
		CommandHandlerTimeout:  cfg.CommandHandlerTimeout,
		Ps2EventHandlerTimeout: cfg.Ps2EventHandlerTimeout,
		Commands: newCommands(
			popLoader,
			worldPopLoader,
			alertsLoader,
		),
		CommandHandlers: newCommandHandlers(
			log,
			popLoader,
			worldPopLoader,
			world_territory_control_loader.NewWorldsTrackerLoader(
				log,
				cfg.BotName,
				platformWorldsTrackers,
			),
			alertsLoader,
			worldAlertsLoader,
			subSettingsLoader,
			platform_character_names_loader.NewCensus(censusClient),
			platform_outfit_tags_loader.NewCensus(censusClient),
			trackable_online_entities_loader.NewCharactersTrackerLoader(
				subSettingsLoader,
				platformCharactersTrackers,
			),
			loaders.NewCachedQueriedLoader(
				log.Logger,
				platform_outfits_loader.NewCensus(censusClient),
				meta.NewPlatformsCache(map[platforms.Platform]containers.MultiKeyedCache[ps2.OutfitId, ps2.Outfit]{
					platforms.PC:     outfits_cache.NewStorage(log, sqlStorage, platforms.PC),
					platforms.PS4_EU: outfits_cache.NewStorage(log, sqlStorage, platforms.PS4_EU),
					platforms.PS4_US: outfits_cache.NewStorage(log, sqlStorage, platforms.PS4_US),
				}),
			),
		),
		SubmitHandlers: newSubmitHandlers(
			sqlStorage,
			censusClient,
			subSettingsLoader,
		),
	}
}
