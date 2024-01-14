package login

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type TrackingManager interface {
	TrackingDiscordChannelIds(characterId string) ([]string, error)
}

func New(tm TrackingManager, charLoader loaders.KeyedLoader[string, ps2.Character]) handlers.Ps2EventHandler[ps2events.PlayerLogin] {
	return func(ctx context.Context, log *slog.Logger, s *discordgo.Session, event ps2events.PlayerLogin) error {
		const op = "handlers.login"
		log = log.With(slog.String("op", op), slog.String("character", event.CharacterID))
		channels, err := tm.TrackingDiscordChannelIds(event.CharacterID)
		if err != nil {
			return fmt.Errorf("%s error getting tracking channels: %w", op, err)
		}
		if len(channels) == 0 {
			return nil
		}
		log.Debug("sending login message for", slog.Int("channels", len(channels)))
		character, err := charLoader.Load(ctx, event.CharacterID)
		if err != nil {
			return fmt.Errorf("%s error getting character: %w", op, err)
		}
		message := render.RenderCharacterLogin(character.Value)
		errors := make([]string, 0, len(channels))
		for _, channel := range channels {
			_, err = s.ChannelMessageSend(channel, message)
			if err != nil {
				log.Error("error sending login message", slog.String("channel", channel), sl.Err(err))
				errors = append(errors, err.Error())
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("%s error sending login message: %s", op, strings.Join(errors, ", "))
		}
		return nil
	}
}
