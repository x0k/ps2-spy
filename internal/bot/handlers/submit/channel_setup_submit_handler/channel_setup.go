package channel_setup_submit_handler

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type Saver interface {
	Save(ctx context.Context, channelId string, settings meta.SubscriptionSettings) error
}

func New(
	charactersLoader loaders.QueriedLoader[[]string, []ps2.CharacterId],
	characterNamesLoader loaders.QueriedLoader[[]ps2.CharacterId, []string],
	outfitsLoader loaders.QueriedLoader[[]string, []ps2.OutfitId],
	outfitTagsLoader loaders.QueriedLoader[[]ps2.OutfitId, []string],
	saver Saver,
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		data := i.ModalSubmitData()
		var err error
		var outfitsIds []ps2.OutfitId
		outfitTagsFromInput := stringsx.SplitAndTrim(
			data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
			",",
		)
		if outfitTagsFromInput[0] != "" {
			outfitsIds, err = outfitsLoader.Load(ctx, outfitTagsFromInput)
			if err != nil {
				return nil, err
			}
		}
		var charIds []ps2.CharacterId
		charNamesFromInput := stringsx.SplitAndTrim(
			data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
			",",
		)
		if charNamesFromInput[0] != "" {
			charIds, err = charactersLoader.Load(ctx, charNamesFromInput)
			if err != nil {
				return nil, err
			}
		}
		err = saver.Save(
			ctx,
			i.ChannelID,
			meta.SubscriptionSettings{
				Outfits:    outfitsIds,
				Characters: charIds,
			},
		)
		if err != nil {
			return nil, err
		}
		outfitTags, err := outfitTagsLoader.Load(ctx, outfitsIds)
		if err != nil {
			return nil, err
		}
		charNames, err := characterNamesLoader.Load(ctx, charIds)
		if err != nil {
			return nil, err
		}
		content := render.RenderSubscriptionsSettingsUpdate(meta.TrackableEntities[[]string, []string]{
			Outfits:    outfitTags,
			Characters: charNames,
		})
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	})
}
