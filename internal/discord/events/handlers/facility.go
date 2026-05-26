package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
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
		return facilityControlHandlerBody(
			ctx, session, messages, outfitLoader, facilityLoader, platform,
			e.Channels, e.Event.FacilityControl,
			func() ps2.OutfitId { return ps2.OutfitId(e.Event.OutfitID) },
			messages.FacilityControl,
		)
	})
}

func NewFacilityLoss(
	m *HandlersManager,
	messages *discord_messages.Messages,
	outfitLoader OutfitLoader,
	facilityLoader FacilityLoader,
	platform ps2_platforms.Platform,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.FacilityLoss,
	) error {
		return facilityControlHandlerBody(
			ctx, session, messages, outfitLoader, facilityLoader, platform,
			e.Channels, e.Event.FacilityControl,
			func() ps2.OutfitId { return e.Event.OldOutfitId },
			messages.FacilityLoss,
		)
	})
}

func facilityControlHandlerBody(
	ctx context.Context,
	session *discordgo.Session,
	messages *discord_messages.Messages,
	outfitLoader OutfitLoader,
	facilityLoader FacilityLoader,
	platform ps2_platforms.Platform,
	channels []discord.Channel,
	facilityEvent ps2events.FacilityControl,
	getOutfitId func() ps2.OutfitId,
	msgFn func(ps2.WorldId, ps2.Outfit, ps2.Facility) discord.Message,
) error {
	return sendSimpleMessage(
		session,
		slicesx.Filter(channels, func(i int) bool {
			return channels[i].OutfitNotifications
		}),
		func() discord.Message {
			facilityId := ps2.FacilityId(facilityEvent.FacilityID)
			facility, err := facilityLoader(ctx, facilityId)
			if err != nil {
				return messages.FacilityLoadError(facilityId, err)
			}
			outfitId := getOutfitId()
			outfitTag, err := outfitLoader(ctx, outfitId)
			if err != nil {
				return messages.OutfitLoadError(outfitId, platform, err)
			}
			worldId := ps2.WorldId(facilityEvent.WorldID)
			return msgFn(worldId, outfitTag, facility)
		}(),
	)
}
