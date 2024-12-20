package discord_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"golang.org/x/text/message"
)

func (m *Messages) channelSettingsForm(
	p *message.Printer,
	channel discord.Channel,
) []discordgo.MessageComponent {
	one := 1
	localeBase, _ := channel.Locale.Base()
	timezoneSelectOptions := m.timezoneOptions(
		p.Sprintf("Default timezone"),
		channel.DefaultTimezone,
	)
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.CHANNEL_LANGUAGE_COMPONENT_CUSTOM_ID,
					Placeholder: p.Sprintf("Language"),
					MinValues:   &one,
					MaxValues:   1,
					Options: []discordgo.SelectMenuOption{
						{
							Label:   p.Sprintf("Language: english"),
							Value:   "en",
							Default: localeBase.String() == "en",
						},
						{
							Label:   p.Sprintf("Language: russian"),
							Value:   "ru",
							Default: localeBase.String() == "ru",
						},
					},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.CHANNEL_CHARACTER_NOTIFICATIONS_COMPONENT_CUSTOM_ID,
					Placeholder: "Character notifications",
					MinValues:   &one,
					MaxValues:   1,
					Options: []discordgo.SelectMenuOption{
						{
							Label:   p.Sprintf("Character notifications: on"),
							Value:   "on",
							Default: channel.CharacterNotifications,
						},
						{
							Label:   p.Sprintf("Character notifications: off"),
							Value:   "off",
							Default: !channel.CharacterNotifications,
						},
					},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.CHANNEL_OUTFIT_NOTIFICATIONS_COMPONENT_CUSTOM_ID,
					Placeholder: "Outfit notifications",
					MinValues:   &one,
					MaxValues:   1,
					Options: []discordgo.SelectMenuOption{
						{
							Label:   p.Sprintf("Outfit notifications: on"),
							Value:   "on",
							Default: channel.OutfitNotifications,
						},
						{
							Label:   p.Sprintf("Outfit notifications: off"),
							Value:   "off",
							Default: !channel.OutfitNotifications,
						},
					},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.CHANNEL_TITLE_UPDATES_COMPONENT_CUSTOM_ID,
					Placeholder: "Title updates",
					MinValues:   &one,
					MaxValues:   1,
					Options: []discordgo.SelectMenuOption{
						{
							Label:   p.Sprintf("Title updates: on"),
							Value:   "on",
							Default: channel.TitleUpdates,
						},
						{
							Label:   p.Sprintf("Title updates: off"),
							Value:   "off",
							Default: !channel.TitleUpdates,
						},
					},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.CHANNEL_DEFAULT_TIMEZONE_COMPONENT_CUSTOM_ID,
					Placeholder: "Default timezone",
					MinValues:   &one,
					MaxValues:   1,
					Options:     timezoneSelectOptions,
				},
			},
		},
	}
}
