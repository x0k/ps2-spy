package discord

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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

type Command struct {
	Cmd            *discordgo.ApplicationCommand
	Handler        InteractionHandler
	SubmitHandlers map[string]InteractionHandler
	// ComponentHandlers map[string]InteractionHandler
}

const (
	SUBSCRIPTION_SETUP_PC_MODAL     = "subscription_setup_pc"
	SUBSCRIPTION_SETUP_PS4_EU_MODAL = "subscription_setup_ps4_eu"
	SUBSCRIPTION_SETUP_PS4_US_MODAL = "subscription_setup_ps4_us"
)

var SUBSCRIPTION_MODAL_CUSTOM_IDS = map[ps2_platforms.Platform]string{
	ps2_platforms.PC:     SUBSCRIPTION_SETUP_PC_MODAL,
	ps2_platforms.PS4_EU: SUBSCRIPTION_SETUP_PS4_EU_MODAL,
	ps2_platforms.PS4_US: SUBSCRIPTION_SETUP_PS4_US_MODAL,
}

func DeferredEphemeralResponse(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) LocalizedEdit) InteractionHandler {
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
		data, customErr := handle(ctx, s, i)(LocaleFromInteraction(i))
		if customErr != nil {
			if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: customErr.Msg,
			}); err != nil {
				log.Error(ctx, "error sending followup message", sl.Err(err))
			}
			return customErr.Err
		}
		_, err = s.InteractionResponseEdit(i.Interaction, data)
		return err
	}
}
