package en_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (m *messages) WorldAlerts(worldName string, worldAlerts meta.Loaded[ps2.Alerts]) (*discordgo.WebhookEdit, *discord.Error) {
	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			RenderWorldDetailedAlerts(worldName, worldAlerts),
		},
	}, nil
}
