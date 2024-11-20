package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"golang.org/x/text/language"
)

type ChannelLanguageSaver = func(ctx context.Context, channelId discord.ChannelId, language language.Tag) error

func NewLanguage(
	messages *discord_messages.Messages,
	languageSaver ChannelLanguageSaver,
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "language",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "язык",
			},
			Description: "Change cannel language",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Сменить язык канала",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionString,
					Name: "language",
					NameLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "язык",
					},
					Description: "Language",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Язык",
					},
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name: "english",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "английский",
							},
							Value: "en",
						},
						{
							Name:  "russian",
							Value: "ru",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "русский",
							},
						},
					},
					Required: true,
				},
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.Edit {
			option := i.ApplicationCommandData().Options[0].StringValue()
			channelId := discord.ChannelId(i.ChannelID)
			lang, err := language.Parse(option)
			if err != nil {
				return messages.ChannelLanguageParseError(channelId, option, err)
			}
			if err := languageSaver(ctx, channelId, lang); err != nil {
				return messages.ChannelLanguageSaveError(channelId, lang, err)
			}
			return messages.ChannelLanguageSaved(channelId, lang)
		}),
	}
}
