package discord_module

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
)

var ErrDuplicateSubmitHandler = errors.New("duplicate submit handler")

func NewSessionService(
	log *logger.Logger,
	fataler module.Fataler,
	session *discordgo.Session,
	commands []*Command,
	commandHandlerTimeout time.Duration,
	removeCommands bool,
) module.Service {
	return module.NewService("discord_session", func(ctx context.Context) error {
		handlers := make(map[string]InteractionHandler, len(commands))
		appCommands := make([]*discordgo.ApplicationCommand, 0, len(commands))
		submitHandlers := make(map[string]InteractionHandler, len(commands))
		for _, command := range commands {
			handlers[command.Cmd.Name] = command.Handler
			if command.SubmitHandlers != nil {
				for name, handler := range command.SubmitHandlers {
					if _, ok := submitHandlers[name]; ok {
						return fmt.Errorf("%w: %s", ErrDuplicateSubmitHandler, name)
					}
					submitHandlers[name] = handler
				}
			}
			appCommands = append(appCommands, command.Cmd)
		}

		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			log.Info(ctx, "logged in as", slog.String("username", s.State.User.Username), slog.String("discriminator", s.State.User.Discriminator))
			log.Info(ctx, "running on", slog.Int("serverCount", len(s.State.Guilds)))
		})
		session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var userId string
			if i.Member != nil {
				userId = i.Member.User.ID
			} else if i.User != nil {
				userId = i.User.ID
			} else {
				userId = i.AppID
			}
			hLog := log.With(
				slog.String("guild_id", i.GuildID),
				slog.String("channel_id", i.ChannelID),
				slog.String("user_id", userId),
			)
			hLog.Debug(ctx, "interaction received", slog.String("type", i.Type.String()))
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				cLog := hLog.With(slog.String("command", i.ApplicationCommandData().Name))
				if handler, ok := handlers[i.ApplicationCommandData().Name]; ok {
					cLog.Debug(ctx, "command received")
					go handler.Run(ctx, cLog, commandHandlerTimeout, s, i)
				} else {
					cLog.Debug(ctx, "command not found")
				}
			case discordgo.InteractionModalSubmit:
				data := i.ModalSubmitData()
				mLog := hLog.With(slog.String("custom_id", data.CustomID))
				if handler, ok := submitHandlers[data.CustomID]; ok {
					mLog.Debug(ctx, "submit received")
					go handler.Run(ctx, mLog, commandHandlerTimeout, s, i)
				} else {
					mLog.Debug(ctx, "submit not found")
				}
			}
		})

		if err := session.Open(); err != nil {
			return err
		}

		registeredCommands, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", appCommands)
		if err != nil {
			return err
		}

		<-ctx.Done()

		if removeCommands {
			for _, v := range registeredCommands {
				if err := session.ApplicationCommandDelete(session.State.User.ID, "", v.ID); err != nil {
					log.Error(ctx, "cannot delete command", slog.String("command", v.Name), sl.Err(err))
				}
			}
		}
		return session.Close()
	})
}
