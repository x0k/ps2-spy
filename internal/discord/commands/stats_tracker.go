package discord_commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type ChannelStatsTrackerTasksLoader = loader.Keyed[discord.ChannelId, []discord.StatsTrackerTask]

func NewStatsTracker(
	messages *discord_messages.Messages,
	statsTracker *stats_tracker.StatsTracker,
	channelStatsTrackerTasksLoader ChannelStatsTrackerTasksLoader,
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
				{
					Type: discordgo.ApplicationCommandOptionSubCommand,
					Name: "schedule",
					NameLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "расписание",
					},
					Description: "Schedule management",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Управление расписанием",
					},
				},
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.ResponseEdit {
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
				if err := statsTracker.StopChannelTracker(ctx, channelId); errors.Is(err, stats_tracker.ErrNoChannelTrackerToStop) {
					return messages.NoChannelTrackerToStop()
				} else if err != nil {
					return messages.StopChannelStatsTrackerError(err)
				}
				return messages.ChannelTrackerWillStoppedSoon()
			case "schedule":
				tasks, err := channelStatsTrackerTasksLoader(ctx, channelId)
				if err != nil {
					return messages.ChannelStatsTrackerTasksLoadError(err)
				}
				return messages.ChannelStatsTrackerSchedule(tasks)
			}
			return messages.InvalidStatsTrackerSubcommand(
				cmd,
				fmt.Errorf("invalid subcommand: %s", cmd),
			)
		}),
	}
}
