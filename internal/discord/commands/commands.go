package discord_commands

import "github.com/x0k/ps2-spy/internal/discord"

func New() []*discord.Command {
	return []*discord.Command{
		NewAbout(),
	}
}
