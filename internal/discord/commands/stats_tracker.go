package discord_commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/expirable_state_container"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type ChannelStatsTrackerTasksLoader = loader.Keyed[discord.ChannelId, []discord.StatsTrackerTask]
type ChannelTimezoneLoader = loader.Keyed[discord.ChannelId, *time.Location]

func NewStatsTracker(
	messages *discord_messages.Messages,
	statsTracker *stats_tracker.StatsTracker,
	channelStatsTrackerTasksLoader ChannelStatsTrackerTasksLoader,
	channelLoader ChannelLoader,
	createTaskStateContainer *expirable_state_container.ExpirableStateContainer[
		discord.ChannelId,
		discord.CreateStatsTrackerTaskState,
	],
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
				channel, err := channelLoader(ctx, channelId)
				if err != nil {
					return discord_messages.ChannelLoadError[discordgo.WebhookEdit](
						channelId,
						err,
					)
				}
				tasks, err := channelStatsTrackerTasksLoader(ctx, channelId)
				if err != nil {
					return messages.ChannelStatsTrackerTasksLoadError(err)
				}
				return messages.ChannelStatsTrackerSchedule(channel, tasks)
			}
			return messages.StatsTrackerInvalidSubcommand(
				cmd,
				fmt.Errorf("invalid subcommand: %s", cmd),
			)
		}),
		ComponentHandlers: map[string]discord.InteractionHandler{
			discord.STATS_TRACKER_TASK_ADD_BUTTON_CUSTOM_ID: discord.MessageUpdate(
				func(
					ctx context.Context,
					s *discordgo.Session,
					i *discordgo.InteractionCreate,
				) discord.Response {
					channelId := discord.ChannelId(i.ChannelID)
					channel, err := channelLoader(ctx, channelId)
					if err != nil {
						return discord_messages.ChannelLoadError[discordgo.InteractionResponseData](
							channelId,
							err,
						)
					}
					state := discord.NewCreateStatsTrackerTaskState(channel.DefaultTimezone)
					createTaskStateContainer.Store(channelId, state)
					return messages.ChannelStatsTrackerAddTaskForm(state)
				},
			),
		},
	}
}
