package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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

type TrackingManager interface {
	ChannelIds(event any) ([]string, error)
}

type Ps2EventHandlerConfig struct {
	Log             *slog.Logger
	Session         *discordgo.Session
	Timeout         time.Duration
	TrackingManager TrackingManager
}

type Ps2EventHandler[E any] func(ctx context.Context, channelIds []string, cfg *Ps2EventHandlerConfig, event E) error

func (handler Ps2EventHandler[E]) Run(ctx context.Context, cfg *Ps2EventHandlerConfig, event E) {
	cfg.Log.Debug("handling", slog.Duration("timeout", cfg.Timeout))
	t := time.Now()
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	err := contextx.Await(ctx, func() error {
		channels, err := cfg.TrackingManager.ChannelIds(event)
		if err != nil {
			return err
		}
		if len(channels) == 0 {
			return nil
		}
		return handler(ctx, channels, cfg, event)
	})
	if err != nil {
		cfg.Log.Error("error handling", sl.Err(err))
	}
	cfg.Log.Debug("handled", slog.Duration("duration", time.Since(t)))
}

func SimpleMessage[E any](handle func(ctx context.Context, cfg *Ps2EventHandlerConfig, event E) (string, error)) Ps2EventHandler[E] {
	return func(ctx context.Context, channelIds []string, cfg *Ps2EventHandlerConfig, event E) error {
		msg, err := handle(ctx, cfg, event)
		if err != nil {
			return err
		}
		errors := make([]string, 0, len(channelIds))
		for _, channel := range channelIds {
			_, err = cfg.Session.ChannelMessageSend(channel, msg)
			if err != nil {
				cfg.Log.Error("error sending login message", slog.String("channel", channel), sl.Err(err))
				errors = append(errors, err.Error())
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error sending login message: %s", strings.Join(errors, ", "))
		}
		return nil
	}
}
