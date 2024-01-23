package channel_setup_command_handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func New(
	settingsLoader loaders.KeyedLoader[meta.SettingsQuery, meta.SubscriptionSettings],
	namesLoader loaders.QueriedLoader[meta.PlatformQuery[ps2.CharacterId], []string],
	outfitTagsLoader loaders.QueriedLoader[meta.PlatformQuery[ps2.OutfitId], []string],
) handlers.InteractionHandler {
	return handlers.ShowModal(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, *handlers.Error) {
		const op = "bot.handlers.command.channel_setup_command_handler"
		platform := platforms.Platform(i.ApplicationCommandData().Options[0].Name)
		settings, err := settingsLoader.Load(ctx, meta.SettingsQuery{
			ChannelId: meta.ChannelId(i.ChannelID),
			Platform:  platform,
		})
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to get settings",
				Err: fmt.Errorf("%s getting settings: %w", op, err),
			}
		}
		tags, err := outfitTagsLoader.Load(ctx, meta.PlatformQuery[ps2.OutfitId]{
			Platform: platform,
			Items:    settings.Outfits,
		})
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to get outfit tags",
				Err: fmt.Errorf("%s getting outfit tags: %w", op, err),
			}
		}
		names, err := namesLoader.Load(ctx, meta.PlatformQuery[ps2.CharacterId]{
			Platform: platform,
			Items:    settings.Characters,
		})
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to get character names",
				Err: fmt.Errorf("%s getting character names: %w", op, err),
			}
		}
		customId, ok := handlers.PlatformModals[platform]
		if !ok {
			return nil, &handlers.Error{
				Msg: "Unsupported platform",
				Err: fmt.Errorf("%s unsupported platform: %s", op, platform),
			}
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
