package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

const (
	CHANNEL_SETUP_PC_MODAL     = "channel_setup_pc"
	CHANNEL_SETUP_PS4_EU_MODAL = "channel_setup_ps4_eu"
	CHANNEL_SETUP_PS4_US_MODAL = "channel_setup_ps4_us"
)

var PlatformModals = map[platforms.Platform]string{
	platforms.PC:     CHANNEL_SETUP_PC_MODAL,
	platforms.PS4_EU: CHANNEL_SETUP_PS4_EU_MODAL,
	platforms.PS4_US: CHANNEL_SETUP_PS4_US_MODAL,
}

var ModalsTitles = map[string]string{
	CHANNEL_SETUP_PC_MODAL:     "Subscription Settings (PC)",
	CHANNEL_SETUP_PS4_EU_MODAL: "Subscription Settings (PS4 EU)",
	CHANNEL_SETUP_PS4_US_MODAL: "Subscription Settings (PS4 US)",
}

type Error struct {
	Msg string
	Err error
}

type InteractionHandler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error

func (handler InteractionHandler) Run(ctx context.Context, timeout time.Duration, s *discordgo.Session, i *discordgo.InteractionCreate) {
	const op = "bot.handlers.InteractionHandler.Run"
	var userId string
	if i.Member != nil {
		userId = i.Member.User.ID
	} else {
		userId = i.User.ID
	}
	log := infra.Logger(ctx).With(
		infra.Op(op),
		slog.String("guild_id", i.GuildID),
		slog.String("channel_id", i.ChannelID),
		slog.String("user_id", userId),
	)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Debug("run handler")
	if err := handler(ctx, s, i); err != nil {
		log.Error("error handling", sl.Err(err))
	}
}

func DeferredEphemeralResponse(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, *Error)) InteractionHandler {
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
		data, customErr := handle(ctx, s, i)
		if customErr != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: customErr.Msg,
			})
			return customErr.Err
		}
		_, err = s.InteractionResponseEdit(i.Interaction, data)
		return err
	}
}

func ShowModal(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, *Error)) InteractionHandler {
	return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		data, err := handle(ctx, s, i)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: err.Msg,
			})
			return err.Err
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: data,
		})
	}
}

type Ps2EventHandlerConfig struct {
	Session                     *discordgo.Session
	Timeout                     time.Duration
	EventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId]
}

type Ps2EventHandler[E any] func(ctx context.Context, channelIds []meta.ChannelId, cfg *Ps2EventHandlerConfig, event E) error

func (handler Ps2EventHandler[E]) Run(ctx context.Context, cfg *Ps2EventHandlerConfig, event E) {
	const op = "bot.handlers.Ps2EventHandler.Run"
	log := infra.Logger(ctx).With(infra.Op(op), slog.Any("event", event))
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	channels, err := cfg.EventTrackingChannelsLoader.Load(ctx, event)
	if err != nil {
		log.Error("error loading tracking channels", sl.Err(err))
		return
	}
	if len(channels) == 0 {
		return
	}
	log.Debug("run handler for", slog.Int("channelsCount", len(channels)))
	if err := handler(ctx, channels, cfg, event); err != nil {
		log.Error("error handling", sl.Err(err))
	}
}

// TODO: publisher.Event instead of `any` to use event.Type()
//       in error message.
//       Currently it is too painful cause `ps2events` uses
//		 pointer receivers in `Type()` method.

func SimpleMessage[E any](handle func(ctx context.Context, cfg *Ps2EventHandlerConfig, event E) (string, *Error)) Ps2EventHandler[E] {
	return func(ctx context.Context, channelIds []meta.ChannelId, cfg *Ps2EventHandlerConfig, event E) error {
		const op = "bot.handlers.SimpleMessage"
		log := infra.OpLogger(ctx, op)
		msg, err := handle(ctx, cfg, event)
		if err != nil {
			msg = err.Msg
		}
		if msg == "" {
			return nil
		}
		errors := make([]string, 0, len(channelIds))
		for len(msg) > 0 {
			toSend := msg
			if len(toSend) > 4000 {
				toSend = toSend[:4000]
				lastLineBreak := strings.LastIndexByte(toSend, '\n')
				if lastLineBreak > 0 {
					toSend = toSend[:lastLineBreak]
					msg = msg[lastLineBreak+1:]
				} else {
					lastSpace := strings.LastIndexByte(toSend, ' ')
					if lastSpace > 0 {
						toSend = toSend[:lastSpace]
						msg = msg[lastSpace+1:]
					} else {
						const truncation = "... (truncated)"
						toSend = msg[:4000-len(truncation)] + truncation
						msg = msg[4000-len(truncation):]
					}
				}
			} else {
				msg = ""
			}
			for _, channelId := range channelIds {
				_, err := cfg.Session.ChannelMessageSend(string(channelId), toSend)
				if err != nil {
					log.Error(
						"error sending message",
						slog.String("channel", string(channelId)),
						sl.Err(err),
					)
					errors = append(errors, err.Error())
				}
			}
		}
		if err != nil {
			return fmt.Errorf("%s handling event: %w", op, err.Err)
		}
		if len(errors) > 0 {
			return fmt.Errorf("%s sending messages: %s", op, strings.Join(errors, ", "))
		}
		return nil
	}
}
