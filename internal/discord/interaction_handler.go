package discord

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"golang.org/x/text/message"
)

type InteractionHandler func(
	ctx context.Context,
	log *logger.Logger,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error

func (handler InteractionHandler) Run(
	ctx context.Context,
	log *logger.Logger,
	timeout time.Duration,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Debug(ctx, "run handler")
	if err := handler(ctx, log, s, i); err != nil {
		log.Error(ctx, "error handling", sl.Err(err))
	}
}

type interactionHandler[R any] func(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) R

func ShowModal(handle interactionHandler[Response]) InteractionHandler {
	return func(ctx context.Context, log *logger.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		data, customErr := handle(ctx, s, i)(message.NewPrinter(UserLocale(i)))
		if customErr != nil {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: customErr.Msg,
				},
			}); err != nil {
				log.Error(ctx, "error sending followup message", sl.Err(err))
			}
			return customErr.Err
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: data,
		})
	}
}

func MessageUpdate(handle interactionHandler[Response]) InteractionHandler {
	return func(
		ctx context.Context,
		log *logger.Logger,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) error {
		data, customErr := handle(ctx, s, i)(message.NewPrinter(UserLocale(i)))
		if customErr != nil {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: customErr.Msg,
				},
			}); err != nil {
				log.Error(ctx, "error sending error message", sl.Err(err))
			}
			return customErr.Err
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: data,
		})
	}
}

func DeferredMessageEdit(handle interactionHandler[MessageEdit]) InteractionHandler {
	return func(
		ctx context.Context,
		log *logger.Logger,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		if err != nil {
			return err
		}
		data, customErr := handle(ctx, s, i)(message.NewPrinter(UserLocale(i)))
		if customErr != nil {
			if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: customErr.Msg,
			}); err != nil {
				log.Error(ctx, "error sending followup message", sl.Err(err))
			}
			return customErr.Err
		}
		_, err = s.ChannelMessageEditComplex(data)
		return err
	}
}

func DeferredEphemeralResponse(handle interactionHandler[ResponseEdit]) InteractionHandler {
	return func(
		ctx context.Context,
		log *logger.Logger,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
		data, customErr := handle(ctx, s, i)(message.NewPrinter(UserLocale(i)))
		if customErr != nil {
			if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: customErr.Msg,
			}); err != nil {
				log.Error(ctx, "error sending followup message", sl.Err(err))
			}
			return customErr.Err
		}
		if data == nil {
			return nil
		}
		_, err = s.InteractionResponseEdit(i.Interaction, data)
		return err
	}
}

func DeferredEphemeralFollowUp(handle interactionHandler[FollowUp]) InteractionHandler {
	return func(ctx context.Context, log *logger.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
		data, customErr := handle(ctx, s, i)(message.NewPrinter(UserLocale(i)))
		if customErr != nil {
			if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: customErr.Msg,
			}); err != nil {
				log.Error(ctx, "error sending followup message", sl.Err(err))
			}
			return customErr.Err
		}
		if data == nil {
			return nil
		}
		_, err = s.FollowupMessageCreate(i.Interaction, false, data)
		return err
	}
}
