package discord_messages

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"golang.org/x/text/message"
)

func scheduleNotes(p *message.Printer, loc *time.Location) string {
	return p.Sprintf(
		`Notes:
- The time is specified in the time zone %q. This can be changed in the channel settings;
- The maximum amount of tasks per channel is %d;
- You can edit a task by clicking on it;
- The “Remove” button deletes immediately without confirmation.
`,
		loc.String(),
		discord.MAX_AMOUNT_OF_TASKS_PER_CHANNEL,
	)
}

const pageSize = 4

func statsTrackerScheduleEditForm(
	p *message.Printer,
	timezone *time.Location,
	tasks []discord.StatsTrackerTask,
	zeroIndexedPage int,
) []discordgo.MessageComponent {
	localTasks := make([]discord.StatsTrackerTaskState, 0, len(tasks))
	for _, t := range tasks {
		localTasks = append(localTasks, discord.NewUpdateStatsTrackerTaskState(
			t, timezone,
		))
	}
	slices.SortFunc(localTasks, func(a, b discord.StatsTrackerTaskState) int {
		w := a.LocalWeekdays[0] - b.LocalWeekdays[0]
		if w != 0 {
			return int(w)
		}
		h := (a.LocalStartHour - b.LocalStartHour)
		if h != 0 {
			return h
		}
		return a.LocalStartMin - b.LocalStartMin
	})
	if len(localTasks) > pageSize {
		shift := zeroIndexedPage * pageSize
		localTasks = localTasks[shift:min(shift+pageSize, len(localTasks))]
	}
	rows := make([]discordgo.MessageComponent, 0, len(localTasks)+1)
	for _, t := range localTasks {
		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: discord.NewStatsTrackerTaskEditButtonCustomId(t.TaskId),
					Label: fmt.Sprintf(
						"%s, %02d:%02d, %s",
						renderWeekday(p, t.LocalWeekdays[0]),
						t.LocalStartHour,
						t.LocalStartMin,
						renderDuration(p, t.Duration),
					),
					Style: discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: discord.NewStatsTrackerTaskRemoveButtonCustomId(t.TaskId),
					Label:    p.Sprintf("Remove"),
					Style:    discordgo.DangerButton,
				},
			},
		})
	}
	addButton := discordgo.Button{
		CustomID: discord.STATS_TRACKER_TASKS_ADD_BUTTON_CUSTOM_ID,
		Label:    p.Sprintf("Add new task"),
		Style:    discordgo.PrimaryButton,
	}
	lastRow := []discordgo.MessageComponent{addButton}
	if len(tasks) > 4 {
		if zeroIndexedPage > 0 {
			lastRow = []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: discord.NewStatsTrackerTaskPageButtonCustomId(zeroIndexedPage - 1),
					Label:    p.Sprintf("Previous"),
					Style:    discordgo.SecondaryButton,
				},
				addButton,
			}
		}
		if zeroIndexedPage < int(math.Ceil(float64(len(tasks))/float64(pageSize)))-1 {
			lastRow = append(lastRow, discordgo.Button{
				CustomID: discord.NewStatsTrackerTaskPageButtonCustomId(zeroIndexedPage + 1),
				Label:    p.Sprintf("Next"),
				Style:    discordgo.SecondaryButton,
			})
		}
	}
	rows = append(rows, discordgo.ActionsRow{
		Components: lastRow,
	})
	return rows
}

func hourPickerOptions(p *message.Printer, selectedHour int) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, 24)
	for i := 0; i < 24; i++ {
		options = append(options, discordgo.SelectMenuOption{
			Label:   p.Sprintf("Starting hour: %d", i),
			Value:   strconv.Itoa(i),
			Default: i == selectedHour,
		})
	}
	return options
}

func minutePickerOptions(p *message.Printer, selectedMinute int) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, 6)
	for i := 0; i < 60; i += 10 {
		options = append(options, discordgo.SelectMenuOption{
			Label:   p.Sprintf("Starting minute: %d", i),
			Value:   strconv.Itoa(i),
			Default: i == selectedMinute,
		})
	}
	return options
}

func durationPickerOptions(p *message.Printer, selectedDuration time.Duration) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, 6)
	for i := 0 * time.Minute; i <= 4*time.Hour; i += 30 * time.Minute {
		options = append(options, discordgo.SelectMenuOption{
			Label:   p.Sprintf("Duration: %s", renderDuration(p, i)),
			Value:   i.String(),
			Default: i == selectedDuration,
		})
	}
	return options
}

func (m *Messages) statsTrackerCreateTaskForm(
	p *message.Printer,
	s discord.StatsTrackerTaskState,
) []discordgo.MessageComponent {
	one := 1
	weekdayOptions := make([]discordgo.SelectMenuOption, 0, 7)
	for i := time.Sunday; i <= time.Saturday; i++ {
		weekdayOptions = append(weekdayOptions, discordgo.SelectMenuOption{
			Label:   renderWeekday(p, i),
			Value:   strconv.Itoa(int(i)),
			Default: slices.Contains(s.LocalWeekdays, i),
		})
	}
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.STATS_TRACKER_TASK_WEEKDAYS_SELECTOR_CUSTOM_ID,
					Placeholder: "Weekdays",
					MinValues:   &one,
					MaxValues:   7,
					Options:     weekdayOptions,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.STATS_TRACKER_TASK_START_HOUR_SELECTOR_CUSTOM_ID,
					Placeholder: "Starting hour",
					MinValues:   &one,
					MaxValues:   1,
					Options:     hourPickerOptions(p, s.LocalStartHour),
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.STATS_TRACKER_TASK_START_MINUTE_SELECTOR_CUSTOM_ID,
					Placeholder: "Starting minute",
					MinValues:   &one,
					MaxValues:   1,
					Options:     minutePickerOptions(p, s.LocalStartMin),
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.STATS_TRACKER_TASK_DURATION_SELECTOR_CUSTOM_ID,
					Placeholder: "Duration",
					MinValues:   &one,
					MaxValues:   1,
					Options:     durationPickerOptions(p, s.Duration),
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: s.SubmitButtonId,
					Label:    p.Sprintf("Submit"),
					Style:    discordgo.SuccessButton,
				},
				discordgo.Button{
					CustomID: discord.STATS_TRACKER_TASK_CANCEL_BUTTON_CUSTOM_ID,
					Label:    p.Sprintf("Cancel"),
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
}
