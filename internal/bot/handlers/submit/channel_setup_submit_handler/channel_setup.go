package channel_setup_submit_handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
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
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		data := i.ModalSubmitData()
		outfits := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		chars := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		charIds, err := characterIdsLoader.Load(ctx, stringsx.SplitAndTrim(chars, ","))
		if err != nil {
			return nil, err
		}
		realOutfits, err := outfitTagsLoader.Load(ctx, stringsx.SplitAndTrim(outfits, ","))
		if err != nil {
			return nil, err
		}
		err = saver.Save(
			ctx,
			i.ChannelID,
			meta.SubscriptionSettings{
				Outfits:    realOutfits,
				Characters: charIds,
			},
		)
		if err != nil {
			return nil, err
		}
		names, err := characterNamesLoader.Load(ctx, charIds)
		if err != nil {
			return nil, err
		}
		content := fmt.Sprintf(
			"Settings are updated.\n\n**Outfits:**\n%s\n\n**Characters:**\n%s",
			strings.Join(realOutfits, ", "),
			strings.Join(names, ", "),
		)
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	})
}
