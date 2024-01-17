package channelsetup

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
)

func New(
	settingsLoader loaders.KeyedLoader[[2]string, meta.SubscriptionSettings],
	namesLoader loaders.QueriedLoader[[]string, []string],
	outfitTagsLoader loaders.QueriedLoader[[]string, []string],
) handlers.InteractionHandler {
	return handlers.ShowModal(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		platform := i.ApplicationCommandData().Options[0].Name
		settings, err := settingsLoader.Load(ctx, [2]string{i.ChannelID, platform})
		if err != nil {
			return nil, err
		}
		tags, err := outfitTagsLoader.Load(ctx, settings.Outfits)
		if err != nil {
			return nil, err
		}
		names, err := namesLoader.Load(ctx, settings.Characters)
		if err != nil {
			return nil, err
		}
		customId, ok := handlers.PlatformModals[platform]
		if !ok {
			return nil, fmt.Errorf("unknown platform: %q", platform)
		}
		return &discordgo.InteractionResponseData{
			CustomID: customId,
			Title:    handlers.ModalsTitles[customId],
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "outfits",
							Label:       "Which outfits do you want to track?",
							Placeholder: "Enter the outfit tags separated by comma",
							Style:       discordgo.TextInputShort,
							Value:       strings.Join(tags, ", "),
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "characters",
							Label:       "Which characters do you want to track?",
							Placeholder: "Enter the character names separated by comma",
							Style:       discordgo.TextInputParagraph,
							Value:       strings.Join(names, ", "),
						},
					},
				},
			},
		}, nil
	})
}
