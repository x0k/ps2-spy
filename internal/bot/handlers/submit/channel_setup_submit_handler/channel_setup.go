package channel_setup_submit_handler

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
)

type Saver interface {
	Save(ctx context.Context, channelId string, settings meta.SubscriptionSettings) error
}

func New(
	characterIdsLoader loaders.QueriedLoader[[]string, []string],
	characterNamesLoader loaders.QueriedLoader[[]string, []string],
	outfitTagsLoader loaders.QueriedLoader[[]string, []string],
	saver Saver,
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		data := i.ModalSubmitData()
		var err error
		var outfitsTags []string
		outfitTagsFromInput := stringsx.SplitAndTrim(
			data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
			",",
		)
		if outfitTagsFromInput[0] != "" {
			outfitsTags, err = outfitTagsLoader.Load(ctx, outfitTagsFromInput)
			if err != nil {
				return nil, err
			}
		}
		var charIds []string
		charNamesFromInput := stringsx.SplitAndTrim(
			data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
			",",
		)
		if charNamesFromInput[0] != "" {
			charIds, err = characterIdsLoader.Load(ctx, charNamesFromInput)
			if err != nil {
				return nil, err
			}
		}
		err = saver.Save(
			ctx,
			i.ChannelID,
			meta.SubscriptionSettings{
				Outfits:    outfitsTags,
				Characters: charIds,
			},
		)
		if err != nil {
			return nil, err
		}
		charNames, err := characterNamesLoader.Load(ctx, charIds)
		if err != nil {
			return nil, err
		}
		content := render.RenderSubscriptionsSettingsUpdate(meta.SubscriptionSettings{
			Outfits:    outfitsTags,
			Characters: charNames,
		})
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	})
}
