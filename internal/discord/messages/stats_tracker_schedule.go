package discord_messages

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/timex"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"golang.org/x/text/message"
)

func (m *Messages) scheduleNotes(p *message.Printer, loc *time.Location) string {
	return p.Sprintf(
		`Notes:
- The time is specified in the time zone %q. This can be changed in the channel settings;
- The maximum amount of tasks per channel is %d;
- You can edit a task by clicking on it;
- The “Remove” button deletes immediately without confirmation.
`,
		loc.String(),
		m.maxTrackingTasks,
	)
}

const pageSize = 4

func newStatsTrackerTaskPageButtonCustomId(
	page int,
) string {
	return discord.STATS_TRACKER_TASKS_PAGE_BUTTON_CUSTOM_ID + discord.CUSTOM_ID_SEPARATOR +
		strconv.Itoa(page)
}

func CustomIdToPage(customId string) (int, error) {
	return strconv.Atoi(
		customId[len(discord.STATS_TRACKER_TASKS_PAGE_BUTTON_CUSTOM_ID)+len(discord.CUSTOM_ID_SEPARATOR):],
	)
}

func newStatsTrackerTaskEditButtonCustomId(
	id stats_tracker.TaskId,
) string {
	return discord.STATS_TRACKER_TASKS_EDIT_BUTTON_CUSTOM_ID + discord.CUSTOM_ID_SEPARATOR +
		strconv.FormatInt(int64(id), 10)
}

func CustomIdToTaskIdToEdit(customId string) (stats_tracker.TaskId, error) {
	v, err := strconv.ParseInt(
		customId[len(discord.STATS_TRACKER_TASKS_EDIT_BUTTON_CUSTOM_ID)+len(discord.CUSTOM_ID_SEPARATOR):],
		10,
		64,
	)
	return stats_tracker.TaskId(v), err
}

func newStatsTrackerTaskRemoveButtonCustomId(
	id stats_tracker.TaskId,
) string {
	return discord.STATS_TRACKER_TASKS_REMOVE_BUTTON_CUSTOM_ID + discord.CUSTOM_ID_SEPARATOR +
		strconv.FormatInt(int64(id), 10)
}

func CustomIdToTaskIdToRemove(customId string) (stats_tracker.TaskId, error) {
	v, err := strconv.ParseInt(
		customId[len(discord.STATS_TRACKER_TASKS_REMOVE_BUTTON_CUSTOM_ID)+len(discord.CUSTOM_ID_SEPARATOR):],
		10,
		64,
	)
	return stats_tracker.TaskId(v), err
}

func newLocalTasks(
	tasks []stats_tracker.Task, timezone *time.Location,
) []stats_tracker.CreateOrUpdateTask {
	localTasks := make([]stats_tracker.CreateOrUpdateTask, 0, len(tasks))
	for _, t := range tasks {
		localTasks = append(localTasks, stats_tracker.NewUpdateTask(t, timezone))
	}
	slices.SortFunc(localTasks, func(a, b stats_tracker.CreateOrUpdateTask) int {
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
	return localTasks
}

func statsTrackerScheduleEditForm(
	p *message.Printer,
	localTasks []stats_tracker.CreateOrUpdateTask,
	zeroIndexedPage int,
) []discordgo.MessageComponent {
	if len(localTasks) > pageSize {
		shift := zeroIndexedPage * pageSize
		localTasks = localTasks[shift:min(shift+pageSize, len(localTasks))]
	}
	rows := make([]discordgo.MessageComponent, 0, len(localTasks)+1)
	for _, t := range localTasks {
		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: newStatsTrackerTaskEditButtonCustomId(t.Id),
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
					CustomID: newStatsTrackerTaskRemoveButtonCustomId(t.Id),
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
	if len(localTasks) > 4 {
		if zeroIndexedPage > 0 {
			lastRow = []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: newStatsTrackerTaskPageButtonCustomId(zeroIndexedPage - 1),
					Label:    p.Sprintf("Previous"),
					Style:    discordgo.SecondaryButton,
				},
				addButton,
			}
		}
		if zeroIndexedPage < int(math.Ceil(float64(len(localTasks))/float64(pageSize)))-1 {
			lastRow = append(lastRow, discordgo.Button{
				CustomID: newStatsTrackerTaskPageButtonCustomId(zeroIndexedPage + 1),
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

func renderStatsTrackerSchedule(
	p *message.Printer,
	localTasks []stats_tracker.CreateOrUpdateTask,
) string {
	sb := strings.Builder{}
	sb.WriteString(p.Sprintf("Schedule:"))
	if len(localTasks) == 0 {
		sb.WriteString(p.Sprintf("\n- No tasks were found"))
		return sb.String()
	}
	for _, t := range localTasks {
		sb.WriteString(p.Sprintf(
			"\n- %s, %02d:%02d, %s",
			renderWeekday(p, t.LocalWeekdays[0]),
			t.LocalStartHour,
			t.LocalStartMin,
			renderDuration(p, t.Duration),
		))
	}
	return sb.String()
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

func (m *Messages) durationPickerOptions(p *message.Printer, selectedDuration time.Duration) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, 6)
	for i := 0 * time.Minute; i <= m.maxTrackingDuration; i += 30 * time.Minute {
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
	s discord.FormState[stats_tracker.CreateOrUpdateTask],
) []discordgo.MessageComponent {
	one := 1
	weekdayOptions := make([]discordgo.SelectMenuOption, 0, 7)
	for i := time.Sunday; i <= time.Saturday; i++ {
		weekdayOptions = append(weekdayOptions, discordgo.SelectMenuOption{
			Label:   renderWeekday(p, i),
			Value:   strconv.Itoa(int(i)),
			Default: slices.Contains(s.Data.LocalWeekdays, i),
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
					Options:     hourPickerOptions(p, s.Data.LocalStartHour),
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
					Options:     minutePickerOptions(p, s.Data.LocalStartMin),
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
					Options:     m.durationPickerOptions(p, s.Data.Duration),
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

func renderTaskFormError(p *message.Printer, err error) string {
	var m stats_tracker.ErrMaxTooManyTasksPerChannel
	if errors.As(err, &m) {
		return p.Sprintf(
			"Max amount of tasks per channel is %d, got %d",
			m.Max, m.Got,
		)
	}
	var d stats_tracker.ErrTaskDurationTooLong
	if errors.As(err, &d) {
		return p.Sprintf(
			"Duration too long: expected max %s got %s",
			renderDuration(p, d.MaxDuration),
			renderDuration(p, d.GotDuration),
		)
	}
	var o stats_tracker.ErrOverlappingTasks
	if errors.As(err, &o) {
		t := o.Tasks[0]
		offset := timex.LocationToOffset(o.Timezone)
		localTime := t.UtcStartTime + offset
		tStartWeekday, tStartTime := timex.NormalizeDate(t.UtcStartWeekday, localTime)
		tEndWeekday, tEndTime := timex.NormalizeDate(t.UtcStartWeekday, localTime+o.Duration)
		tDuration := tEndTime - tStartTime
		if tStartWeekday != tEndWeekday {
			tDuration += 24 * time.Hour
		}
		return p.Sprintf(
			"Your task (%s, %s, %s) overlaps with existing task (%s, %s, %s)",
			renderWeekday(p, o.LocalWeekday),
			renderDurationAsTime(o.LocalStartTime),
			renderDuration(p, o.Duration),
			renderWeekday(p, tStartWeekday),
			renderDurationAsTime(tStartTime),
			renderDuration(p, tDuration),
		)
	}
	return p.Sprintf("Something went wrong")
}
