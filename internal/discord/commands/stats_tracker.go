package discord_commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type ChannelStatsTrackerTasksLoader = loader.Keyed[discord.ChannelId, []discord.StatsTrackerTask]
type ChannelTimezoneLoader = loader.Keyed[discord.ChannelId, *time.Location]
type ChannelStatsTrackerTaskCreator = func(
	context.Context, discord.ChannelId, discord.StatsTrackerTaskState,
) error
type ChannelStatsTrackerTaskRemover = func(
	context.Context, discord.ChannelId, discord.StatsTrackerTaskId,
) error
type StatsTrackerTaskLoader = loader.Keyed[discord.StatsTrackerTaskId, discord.StatsTrackerTask]
type ChannelStatsTrackerTaskUpdater = func(
	context.Context, discord.ChannelId, discord.StatsTrackerTaskState,
) error

func newStateId(i *discordgo.InteractionCreate) discord.ChannelAndUserIds {
	var userId string
	if i.Member != nil {
		userId = i.Member.User.ID
	} else if i.User != nil {
		userId = i.User.ID
	} else {
		userId = i.AppID
	}
	return discord.NewChannelAndUserId(
		discord.ChannelId(i.Interaction.ChannelID),
		discord.UserId(userId),
	)
}

func NewStatsTracker(
	log *logger.Logger,
	messages *discord_messages.Messages,
	statsTracker *stats_tracker.StatsTracker,
	channelStatsTrackerTasksLoader ChannelStatsTrackerTasksLoader,
	channelLoader ChannelLoader,
	taskStateContainer *containers.ExpirableState[
		discord.ChannelAndUserIds,
		discord.StatsTrackerTaskState,
	],
	statsTrackerTaskCreator ChannelStatsTrackerTaskCreator,
	channelStatsTrackerTaskRemover ChannelStatsTrackerTaskRemover,
	statsTrackerTaskLoader StatsTrackerTaskLoader,
	channelStatsTrackerTaskUpdater ChannelStatsTrackerTaskUpdater,
) *discord.Command {
	newCreateFormHandler := func(
		stateUpdater func(*discordgo.InteractionCreate, discord.StatsTrackerTaskState) (discord.StatsTrackerTaskState, error),
	) discord.InteractionHandler {
		return discord.MessageUpdate(func(
			ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
		) discord.Response {
			stateId := newStateId(i)
			state, ok := taskStateContainer.Pop(stateId)
			if !ok {
				return messages.ChannelStatsTrackerTaskStateNotFound(
					fmt.Errorf("failed to find state %q: %w", stateId, shared.ErrNotFound),
				)
			}
			state, err := stateUpdater(i, state)
			if err != nil {
				return messages.FieldValueExtractError(err)
			}
			taskStateContainer.Store(stateId, state)
			return messages.StatsTrackerTaskForm(state, nil)
		})
	}
	updatedSchedule := func(
		ctx context.Context, i *discordgo.InteractionCreate, zeroIndexedPage int,
	) discord.Response {
		channelId := discord.ChannelId(i.Interaction.ChannelID)
		channel, err := channelLoader(ctx, channelId)
		if err != nil {
			return discord_messages.ChannelLoadError[discordgo.InteractionResponseData](
				channelId,
				err,
			)
		}
		tasks, err := channelStatsTrackerTasksLoader(ctx, channelId)
		if err != nil {
			return discord_messages.ChannelStatsTrackerTasksLoadError[discordgo.InteractionResponseData](
				err,
			)
		}
		return messages.StatsTrackerScheduleUpdated(channel, tasks, zeroIndexedPage)
	}
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name:        "stats-tracker",
			Description: "Stats tracker management",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Управление трекером статистики",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "start",
					Description: "Start stats tracker",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Запустить трекер статистики",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "stop",
					Description: "Stop stats tracker",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Остановить трекер статистики",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "schedule",
					Description: "Schedule management",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Управление расписанием",
					},
				},
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
		) discord.ResponseEdit {
			option := i.ApplicationCommandData().Options[0]
			cmd := option.Name
			channelId := discord.ChannelId(i.ChannelID)
			switch cmd {
			case "start":
				if !discord.IsChannelsManagerOrDM(i) {
					return discord_messages.MissingPermissionError[discordgo.WebhookEdit]()
				}
				if err := statsTracker.StartChannelTracker(ctx, channelId); errors.Is(err, stats_tracker.ErrNothingToTrack) {
					return messages.NothingToTrack()
				} else if err != nil {
					return messages.StartChannelStatsTrackerError(err)
				}
				return messages.ChannelTrackerWillStartedSoon()
			case "stop":
				if !discord.IsChannelsManagerOrDM(i) {
					return discord_messages.MissingPermissionError[discordgo.WebhookEdit]()
				}
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
					return discord_messages.ChannelStatsTrackerTasksLoadError[discordgo.WebhookEdit](
						err,
					)
				}
				if discord.IsChannelsManagerOrDM(i) {
					return messages.StatsTrackerScheduleEditForm(channel, tasks)
				}
				return messages.StatsTrackerSchedule(channel, tasks)
			}
			return messages.StatsTrackerInvalidSubcommand(
				cmd,
				fmt.Errorf("invalid subcommand: %s", cmd),
			)
		}),
		ComponentHandlers: map[string]discord.InteractionHandler{
			discord.STATS_TRACKER_TASKS_EDIT_BUTTON_CUSTOM_ID: discord.MessageUpdate(func(
				ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
			) discord.Response {
				channelId := discord.ChannelId(i.ChannelID)
				channel, err := channelLoader(ctx, channelId)
				if err != nil {
					return discord_messages.ChannelLoadError[discordgo.InteractionResponseData](
						channelId,
						err,
					)
				}
				taskId, err := discord.CustomIdToTaskIdToEdit(i.MessageComponentData().CustomID)
				if err != nil {
					return messages.FieldValueExtractError(err)
				}
				task, err := statsTrackerTaskLoader(ctx, taskId)
				if err != nil {
					return messages.StatsTrackerTaskLoadError(err)
				}
				state := discord.NewUpdateStatsTrackerTaskState(task, channel.DefaultTimezone)
				stateId := newStateId(i)
				taskStateContainer.Store(stateId, state)
				return messages.StatsTrackerTaskForm(state, nil)
			}),
			discord.STATS_TRACKER_TASKS_REMOVE_BUTTON_CUSTOM_ID: discord.MessageUpdate(func(
				ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
			) discord.Response {
				channelId := discord.ChannelId(i.ChannelID)
				taskId, err := discord.CustomIdToTaskIdToRemove(i.MessageComponentData().CustomID)
				if err != nil {
					return messages.FieldValueExtractError(err)
				}
				if err := channelStatsTrackerTaskRemover(ctx, channelId, taskId); err != nil {
					return messages.ChannelStatsTrackerTaskRemoveError(err)
				}
				return updatedSchedule(ctx, i, 0)
			}),
			discord.STATS_TRACKER_TASKS_PAGE_BUTTON_CUSTOM_ID: discord.MessageUpdate(func(
				ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
			) discord.Response {
				page, err := discord.CustomIdToPage(i.MessageComponentData().CustomID)
				if err != nil {
					return messages.FieldValueExtractError(err)
				}
				return updatedSchedule(ctx, i, page)
			}),
			discord.STATS_TRACKER_TASKS_ADD_BUTTON_CUSTOM_ID: discord.MessageUpdate(
				func(
					ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
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
					stateId := newStateId(i)
					taskStateContainer.Store(stateId, state)
					return messages.StatsTrackerTaskForm(state, nil)
				},
			),
			discord.STATS_TRACKER_TASK_WEEKDAYS_SELECTOR_CUSTOM_ID: newCreateFormHandler(
				func(i *discordgo.InteractionCreate, state discord.StatsTrackerTaskState) (discord.StatsTrackerTaskState, error) {
					weekdays := make([]time.Weekday, 0, len(i.MessageComponentData().Values))
					for _, v := range i.MessageComponentData().Values {
						weekday, err := strconv.Atoi(v)
						if err != nil {
							return state, err
						}
						if weekday < 0 || weekday > 6 {
							return state, fmt.Errorf("invalid weekday: %d", weekday)
						}
						weekdays = append(weekdays, time.Weekday(weekday))
					}
					state.LocalWeekdays = weekdays
					return state, nil
				},
			),
			discord.STATS_TRACKER_TASK_START_HOUR_SELECTOR_CUSTOM_ID: newCreateFormHandler(
				func(ic *discordgo.InteractionCreate, state discord.StatsTrackerTaskState) (discord.StatsTrackerTaskState, error) {
					h, err := strconv.Atoi(ic.MessageComponentData().Values[0])
					if err != nil {
						return state, err
					}
					if h < 0 || h > 23 {
						return state, fmt.Errorf("invalid hour: %d", h)
					}
					state.LocalStartHour = h
					return state, nil
				},
			),
			discord.STATS_TRACKER_TASK_START_MINUTE_SELECTOR_CUSTOM_ID: newCreateFormHandler(
				func(ic *discordgo.InteractionCreate, state discord.StatsTrackerTaskState) (discord.StatsTrackerTaskState, error) {
					m, err := strconv.Atoi(ic.MessageComponentData().Values[0])
					if err != nil {
						return state, err
					}
					if m%10 != 0 || m > 59 {
						return state, fmt.Errorf("invalid minute: %d", m)
					}
					state.LocalStartMin = m
					return state, nil
				},
			),
			discord.STATS_TRACKER_TASK_DURATION_SELECTOR_CUSTOM_ID: newCreateFormHandler(
				func(ic *discordgo.InteractionCreate, state discord.StatsTrackerTaskState) (discord.StatsTrackerTaskState, error) {
					d, err := time.ParseDuration(ic.MessageComponentData().Values[0])
					if err != nil {
						return state, err
					}
					state.Duration = d
					return state, nil
				},
			),
			discord.STATS_TRACKER_TASK_CANCEL_BUTTON_CUSTOM_ID: discord.MessageUpdate(func(
				ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
			) discord.Response {
				taskStateContainer.Remove(newStateId(i))
				return updatedSchedule(ctx, i, 0)
			}),
			discord.STATS_TRACKER_TASK_CREATE_SUBMIT_BUTTON_CUSTOM_ID: discord.MessageUpdate(func(
				ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
			) discord.Response {
				stateId := newStateId(i)
				state, ok := taskStateContainer.Pop(stateId)
				if !ok {
					return messages.ChannelStatsTrackerTaskStateNotFound(
						fmt.Errorf("failed to find state %q: %w", stateId, shared.ErrNotFound),
					)
				}
				channelId := discord.ChannelId(i.Interaction.ChannelID)
				err := statsTrackerTaskCreator(ctx, channelId, state)
				if err != nil {
					taskStateContainer.Store(stateId, state)
					log.Debug(ctx, "failed to create task", sl.Err(err))
					return messages.StatsTrackerTaskForm(state, err)
				}
				return updatedSchedule(ctx, i, 0)
			}),
			discord.STATS_TRACKER_TASK_UPDATE_SUBMIT_BUTTON_CUSTOM_ID: discord.MessageUpdate(func(
				ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
			) discord.Response {
				stateId := newStateId(i)
				state, ok := taskStateContainer.Pop(stateId)
				if !ok {
					return messages.ChannelStatsTrackerTaskStateNotFound(
						fmt.Errorf("failed to find state %q: %w", stateId, shared.ErrNotFound),
					)
				}
				channelId := discord.ChannelId(i.Interaction.ChannelID)
				err := channelStatsTrackerTaskUpdater(ctx, channelId, state)
				if err != nil {
					taskStateContainer.Store(stateId, state)
					log.Debug(ctx, "failed to update task", sl.Err(err))
					return messages.StatsTrackerTaskForm(state, err)
				}
				return updatedSchedule(ctx, i, 0)
			}),
		},
	}
}
