package en_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (m *messages) WorldAlertsLoadError(provider string, worldId ps2.WorldId, err error) (*discordgo.WebhookEdit, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Failed to load world alerts for " + string(worldId) + " from " + provider,
		Err: err,
	}
}
