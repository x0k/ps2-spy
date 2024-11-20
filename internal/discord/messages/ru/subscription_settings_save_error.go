package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (m *messages) SubscriptionSettingsSaveError(channelId discord.ChannelId, platform ps2_platforms.Platform, err error) (*discordgo.WebhookEdit, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Ошибка сохранения настроек подписки " + string(channelId) + " (" + string(platform) + ")",
		Err: err,
	}
}
