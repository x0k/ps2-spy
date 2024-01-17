package channelsetup

import (
	"context"
	"fmt"
	"log/slog"

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
	saver Saver,
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		data := i.ModalSubmitData()
		outfits := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		chars := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		charIds, err := characterIdsLoader.Load(ctx, stringsx.SplitAndTrim(chars, ","))
		fmt.Println(chars)
		fmt.Println(stringsx.SplitAndTrim(chars, ","))
		fmt.Println(charIds)
		if err != nil {
			return nil, err
		}
		err = saver.Save(
			ctx,
			i.ChannelID,
			meta.SubscriptionSettings{
				Outfits:    stringsx.SplitAndTrim(outfits, ","),
				Characters: charIds,
			},
		)
		if err != nil {
			return nil, err
		}
		content := fmt.Sprintf(
			"Settings are updated.\n\n**Outfits:**\n%s\n\n**Characters:**\n%s",
			outfits,
			chars,
		)
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	})
}
