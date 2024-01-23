package channel_setup_submit_handler

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type Saver interface {
	Save(ctx context.Context, channelId meta.ChannelId, settings meta.SubscriptionSettings) error
}

func New(
	charactersLoader loaders.QueriedLoader[[]string, []ps2.CharacterId],
	characterNamesLoader loaders.QueriedLoader[[]ps2.CharacterId, []string],
	outfitsLoader loaders.QueriedLoader[[]string, []ps2.OutfitId],
	outfitTagsLoader loaders.QueriedLoader[[]ps2.OutfitId, []string],
	saver Saver,
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(
		ctx context.Context,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) (*discordgo.WebhookEdit, *handlers.Error) {
		const op = "bot.handlers.submit.channel_setup_submit_handler"
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
				return nil, &handlers.Error{
					Msg: "Failed to load outfits",
					Err: fmt.Errorf("%s getting outfit tags: %w", op, err),
				}
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
				return nil, &handlers.Error{
					Msg: "Failed to load characters",
					Err: fmt.Errorf("%s getting character names: %w", op, err),
				}
			}
		}
		err = saver.Save(
			ctx,
			meta.ChannelId(i.ChannelID),
			meta.SubscriptionSettings{
				Outfits:    outfitsIds,
				Characters: charIds,
			},
		)
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to save settings",
				Err: fmt.Errorf("%s saving settings: %w", op, err),
			}
		}
		outfitTags, err := outfitTagsLoader.Load(ctx, outfitsIds)
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Settings are saved, but failed to get outfit tags",
				Err: fmt.Errorf("%s getting outfit tags: %w", op, err),
			}
		}
		charNames, err := characterNamesLoader.Load(ctx, charIds)
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Settings are saved, but failed to get character names",
				Err: fmt.Errorf("%s getting character names: %w", op, err),
			}
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
