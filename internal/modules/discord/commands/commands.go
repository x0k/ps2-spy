package commands

import discord_module "github.com/x0k/ps2-spy/internal/modules/discord"

func New() []*discord_module.Command {
	return []*discord_module.Command{
		NewAbout(),
	}
}
