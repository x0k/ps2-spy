package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/slicesx"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type OutfitLoader = loader.Keyed[ps2.OutfitId, ps2.Outfit]
type FacilityLoader = loader.Keyed[ps2.FacilityId, ps2.Facility]

func NewFacilityControl(
	m *HandlersManager,
	messages *discord_messages.Messages,
	outfitLoader OutfitLoader,
	facilityLoader FacilityLoader,
	platform ps2_platforms.Platform,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.FacilityControl,
	) error {
		return sendSimpleMessage(
			session,
			slicesx.Filter(e.Channels, func(i int) bool {
				return e.Channels[i].OutfitNotifications
			}),
			func() discord.Message {
				facilityId := ps2.FacilityId(e.Event.FacilityID)
				facility, err := facilityLoader(ctx, facilityId)
				if err != nil {
					return messages.FacilityLoadError(facilityId, err)
				}
				outfitId := ps2.OutfitId(e.Event.OutfitID)
				outfitTag, err := outfitLoader(ctx, outfitId)
				if err != nil {
					return messages.OutfitLoadError(outfitId, platform, err)
				}
				worldId := ps2.WorldId(e.Event.WorldID)
				return messages.FacilityControl(worldId, outfitTag, facility)
			}(),
		)
	})
}
