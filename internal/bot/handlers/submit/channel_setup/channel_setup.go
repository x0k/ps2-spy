package channelsetup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	"github.com/x0k/ps2-spy/internal/meta"
)

type Saver interface {
	Save(ctx context.Context, channelId string, settings meta.SubscriptionSettings) error
}

func New(
	saver Saver,
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		data := i.ModalSubmitData()
		outfits := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		chars := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		err := saver.Save(
			ctx,
			i.ChannelID,
			meta.SubscriptionSettings{
				Outfits:    stringsx.SplitAndTrim(outfits, ","),
				Characters: stringsx.SplitAndTrim(chars, ","),
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
