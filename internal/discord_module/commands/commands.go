package commands

import (
	"github.com/x0k/ps2-spy/internal/discord_module"
)

func New() []*discord_module.Command {
	return []*discord_module.Command{
		NewAbout(),
	}
}
