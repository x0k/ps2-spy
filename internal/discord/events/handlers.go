package discord_events

import (
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func NewHandlers(
	log *logger.Logger,
	messages *discord_messages.Messages,
	characterLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character],
	outfitLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit],
	charactersLoaders map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character],
	facilityLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility],
	onlineTrackableEntitiesCountLoader loader.Keyed[discord.ChannelId, int],
	channelTitleUpdater ChannelTitleUpdater,
) map[EventType][]HandlerFactory {
	updateOnlineCountInTitleHandlerFactory := NewUpdateOnlineCountInTitleHandlerFactory(
		log,
		messages,
		onlineTrackableEntitiesCountLoader,
		channelTitleUpdater,
	)
	return map[EventType][]HandlerFactory{
		PlayerLoginType: {
			NewLoginHandlerFactory(
				messages,
				characterLoaders,
			),
			updateOnlineCountInTitleHandlerFactory,
		},
		PlayerLogoutType: {
			NewLogoutHandlerFactory(
				messages,
				characterLoaders,
			),
			updateOnlineCountInTitleHandlerFactory,
		},
		OutfitMembersUpdateType: {
			NewOutfitMembersUpdateHandlerFactory(
				log.With(sl.Component("outfit_members_update_handler_factory")),
				messages,
				outfitLoaders,
				charactersLoaders,
			),
		},
		ChannelLanguageUpdatedType: {
			updateOnlineCountInTitleHandlerFactory,
		},
		FacilityControlType: {
			NewFacilityControlHandlerFactory(
				messages,
				outfitLoaders,
				facilityLoaders,
			),
		},
		FacilityLossType: {
			NewFacilityLossHandlerFactory(
				messages,
				outfitLoaders,
				facilityLoaders,
			),
		},
	}
}
