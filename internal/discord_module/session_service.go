package discord_module

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/module"
)

func NewSessionService(
	session *discordgo.Session,
	fataler module.Fataler,
) module.Hook {
	return module.NewHook("discord_session", func(ctx context.Context) error {
		context.AfterFunc(ctx, func() {
			if err := session.Close(); err != nil {
				fataler.Fatal(ctx, err)
			}
		})
		return session.Open()
	})
}
