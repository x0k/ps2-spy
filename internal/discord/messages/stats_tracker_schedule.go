package discord_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

func statsTrackerScheduleEditForm(tasks []discord.StatsTrackerTask) []discordgo.MessageComponent {
	rows := make([]discordgo.ActionsRow, 0, len(tasks))

	return []discordgo.MessageComponent{}
}
