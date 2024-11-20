package en_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (m *messages) CharacterNamesLoadError(characterIds []ps2.CharacterId, platform ps2_platforms.Platform, err error) (*discordgo.InteractionResponseData, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Failed to load character names for " + string(platform) + " platform",
		Err: err,
	}
}
