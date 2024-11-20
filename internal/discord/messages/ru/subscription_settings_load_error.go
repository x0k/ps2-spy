package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (m *messages) SubscriptionSettingsLoadError(channelId discord.ChannelId, platform ps2_platforms.Platform, err error) (*discordgo.InteractionResponseData, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Ошибка загрузки настроек подписки " + string(channelId) + " (" + string(platform) + ")",
		Err: err,
	}
}
