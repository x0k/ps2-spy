package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/contextx"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

const (
	CHANNEL_SETUP_PC_MODAL     = "channel_setup_pc"
	CHANNEL_SETUP_PS4_EU_MODAL = "channel_setup_ps4_eu"
	CHANNEL_SETUP_PS4_US_MODAL = "channel_setup_ps4_us"
)

var PlatformModals = map[string]string{
	platforms.PC:     CHANNEL_SETUP_PC_MODAL,
	platforms.PS4_EU: CHANNEL_SETUP_PS4_EU_MODAL,
	platforms.PS4_US: CHANNEL_SETUP_PS4_US_MODAL,
}

var ModalsTitles = map[string]string{
	CHANNEL_SETUP_PC_MODAL:     "Subscription Settings (PC)",
	CHANNEL_SETUP_PS4_EU_MODAL: "Subscription Settings (PS4 EU)",
	CHANNEL_SETUP_PS4_US_MODAL: "Subscription Settings (PS4 US)",
}

type InteractionHandler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error

func (handler InteractionHandler) Run(ctx context.Context, timeout time.Duration, s *discordgo.Session, i *discordgo.InteractionCreate) {
	const op = "bot.handlers.InteractionHandler.Run"
	log := infra.OpLogger(ctx, op)
	t := time.Now()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := contextx.Await(ctx, func() error {
		err := handler(ctx, s, i)
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

func DeferredEphemeralResponse(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error)) InteractionHandler {
	return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
		data, err := handle(ctx, s, i)
		if err != nil {
			return err
		}
		_, err = s.InteractionResponseEdit(i.Interaction, data)
		return err
	}
}

func ShowModal(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)) InteractionHandler {
	return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		data, err := handle(ctx, s, i)
		if err != nil {
			return err
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: data,
		})
	}
}

type TrackingManager interface {
	ChannelIds(ctx context.Context, event any) ([]string, error)
}

type Ps2EventHandlerConfig struct {
	Session                     *discordgo.Session
	Timeout                     time.Duration
	EventTrackingChannelsLoader loaders.QueriedLoader[any, []string]
}

type Ps2EventHandler[E any] func(ctx context.Context, channelIds []string, cfg *Ps2EventHandlerConfig, event E) error

func (handler Ps2EventHandler[E]) Run(ctx context.Context, cfg *Ps2EventHandlerConfig, event E) {
	const op = "bot.handlers.Ps2EventHandler.Run"
	log := infra.OpLogger(ctx, op)
	// log.Debug("handling", slog.Duration("timeout", cfg.Timeout))
	// t := time.Now()
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	err := contextx.Await(ctx, func() error {
		// log.Debug("check for tracking channels for", slog.Any("event", event))
		channels, err := cfg.EventTrackingChannelsLoader.Load(ctx, event)
		if err != nil {
			return err
		}
		if len(channels) == 0 {
			return nil
		}
		log.Debug("handling", slog.Any("event", event), slog.Int("channels_count", len(channels)))
		return handler(ctx, channels, cfg, event)
	})
	if err != nil {
		log.Error("error handling", sl.Err(err))
	}
	// log.Debug("handled", slog.Duration("duration", time.Since(t)))
}

func SimpleMessage[E any](handle func(ctx context.Context, cfg *Ps2EventHandlerConfig, event E) (string, error)) Ps2EventHandler[E] {
	return func(ctx context.Context, channelIds []string, cfg *Ps2EventHandlerConfig, event E) error {
		const op = "bot.handlers.SimpleMessage"
		log := infra.OpLogger(ctx, op)
		msg, err := handle(ctx, cfg, event)
		if err != nil {
			return err
		}
		if msg == "" {
			return nil
		}
		errors := make([]string, 0, len(channelIds))
		for _, channel := range channelIds {
			_, err = cfg.Session.ChannelMessageSend(channel, msg)
			if err != nil {
				log.Error("error sending message", slog.String("channel", channel), sl.Err(err))
				errors = append(errors, err.Error())
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error sending message: %s", strings.Join(errors, ", "))
		}
		return nil
	}
}
