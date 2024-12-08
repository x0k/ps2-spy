package discord_commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

func NewStatsTracker(
	messages *discord_messages.Messages,
	statsTracker *stats_tracker.StatsTracker,
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "stats-tracker",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "трекер-статистики",
			},
			Description: "Stats tracker management",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Управление трекером статистики",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionSubCommand,
					Name: "start",
					NameLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "запустить",
					},
					Description: "Start stats tracker",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Запустить трекер статистики",
					},
				},
				{
					Type: discordgo.ApplicationCommandOptionSubCommand,
					Name: "stop",
					NameLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "остановить",
					},
					Description: "Stop stats tracker",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Остановить трекер статистики",
					},
				},
			},
		},
		Handler: discord.DeferredEphemeralEdit(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) discord.Edit {
			option := i.ApplicationCommandData().Options[0]
			cmd := option.Name
			channelId := discord.ChannelId(i.ChannelID)
			switch cmd {
			case "start":
				if err := statsTracker.StartChannelTracker(ctx, channelId); errors.Is(err, stats_tracker.ErrNothingToTrack) {
					return messages.NothingToTrack()
				} else if err != nil {
					return messages.StartChannelStatsTrackerError(err)
				}
				return messages.ChannelTrackerWillStartedSoon()
			case "stop":
				if err := statsTracker.StopChannelTracker(channelId); errors.Is(err, stats_tracker.ErrNoChannelTrackerToStop) {
					return messages.NoChannelTrackerToStop()
				} else if err != nil {
					return messages.StopChannelStatsTrackerError(err)
				}
				return messages.ChannelTrackerWillStoppedSoon()
			}
			return messages.InvalidStatsTrackerSubcommand(
				cmd,
				fmt.Errorf("invalid subcommand: %s", cmd),
			)
		}),
	}
}
