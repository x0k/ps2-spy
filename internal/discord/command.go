package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/message"
)

type Command struct {
	Cmd            *discordgo.ApplicationCommand
	Handler        InteractionHandler
	SubmitHandlers map[string]InteractionHandler
	// ComponentHandlers map[string]InteractionHandler
}

var TRACKING_MODAL_CUSTOM_IDS = map[ps2_platforms.Platform]string{
	ps2_platforms.PC:     "tracking_setup_pc",
	ps2_platforms.PS4_EU: "tracking_setup_ps4_eu",
	ps2_platforms.PS4_US: "tracking_setup_ps4_us",
}

func DeferredEphemeralResponse(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) Edit) InteractionHandler {
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
		data, customErr := handle(ctx, s, i)(message.NewPrinter(LangTagFromInteraction(i)))
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
