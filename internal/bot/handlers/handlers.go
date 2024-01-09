package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/contextx"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

type InteractionHandler func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) error

func (handler InteractionHandler) Run(ctx context.Context, log *slog.Logger, timeout time.Duration, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Debug("handling", slog.Duration("timeout", timeout))
	t := time.Now()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := contextx.Await(ctx, func() error {
		err := handler(ctx, log, s, i)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: err.Error(),
			})
		}
		return err
	})
	if err != nil {
		log.Error("error handling", sl.Err(err))
	}
	log.Debug("handled", slog.Duration("duration", time.Since(t)))
}

func DeferredResponse(handle func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error)) InteractionHandler {
	return func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			return err
		}
		data, err := handle(ctx, log, s, i)
		if err != nil {
			return err
		}
		_, err = s.InteractionResponseEdit(i.Interaction, data)
		return err
	}
}
