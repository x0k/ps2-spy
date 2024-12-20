package discord_messages

import (
	"slices"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"golang.org/x/text/message"
)

func timezoneData(p *message.Printer, loc *time.Location) (string, time.Duration) {
	timezone, offsetInSeconds := time.Now().In(loc).Zone()
	return p.Sprintf(
			"The time is in %q time zone, you can change this in the channel settings",
			timezone,
		),
		time.Duration(offsetInSeconds) * time.Second
}

func localDate(weekday time.Weekday, utcTime time.Duration, offset time.Duration) (time.Weekday, time.Duration) {
	localTime := utcTime + offset
	if localTime < 0 {
		if weekday == time.Sunday {
			weekday = time.Saturday
		} else {
			weekday--
		}
		localTime += 24 * time.Hour
	} else if localTime >= 24*time.Hour {
		if weekday == time.Saturday {
			weekday = time.Sunday
		} else {
			weekday++
		}
		localTime -= 24 * time.Hour
	}
	return weekday, localTime
}

type localStatsTrackerTask struct {
	id           discord.StatsTrackerTaskId
	startWeekday time.Weekday
	startTime    time.Duration
	endWeekday   time.Weekday
	endTime      time.Duration
}

func statsTrackerScheduleEditForm(
	p *message.Printer,
	tasks []discord.StatsTrackerTask,
	offset time.Duration,
) []discordgo.MessageComponent {
	localTasks := make([]localStatsTrackerTask, 0, len(tasks))
	for _, t := range tasks {
		startWeekday, startTime := localDate(t.UtcWeekday, t.UtcStartTime, offset)
		endWeekday, endTime := localDate(t.UtcWeekday, t.UtcEndTime, offset)
		localTasks = append(localTasks, localStatsTrackerTask{
			id:           t.Id,
			startWeekday: startWeekday,
			startTime:    startTime,
			endWeekday:   endWeekday,
			endTime:      endTime,
		})
	}
	slices.SortFunc(localTasks, func(a, b localStatsTrackerTask) int {
		w := a.startWeekday - b.startWeekday
		if w != 0 {
			return int(w)
		}
		return int(a.startTime - b.startTime)
	})
	rows := make([]discordgo.MessageComponent, 0, len(localTasks)+1)
	for _, t := range localTasks {
		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: discord.NewStatsTrackerTaskEditButtonCustomId(t.id),
					Label:    p.Sprintf("Edit"),
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: discord.NewStatsTrackerTaskRemoveButtonCustomId(t.id),
					Label:    p.Sprintf("Remove"),
					Style:    discordgo.DangerButton,
				},
			},
		})
	}
	rows = append(rows, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: discord.STATS_TRACKER_TASK_ADD_BUTTON_CUSTOM_ID,
				Label:    p.Sprintf("Add new task"),
				Style:    discordgo.PrimaryButton,
			},
		},
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
			Value:   strconv.FormatInt(int64(i), 10),
			Default: i == selectedDuration,
		})
	}
	return options
}

func (m *Messages) statsTrackerScheduleAddForm(
	p *message.Printer,
	s discord.CreateStatsTrackerTaskState,
) []discordgo.MessageComponent {
	one := 1
	selectedHour := int(s.LocalStartTime / time.Hour)
	selectedMinute := int((s.LocalStartTime % time.Hour) / time.Minute)
	timezoneOpts := m.timezoneOptions(
		p.Sprintf("Timezone"),
		s.Timezone,
	)
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.STATS_TRACKER_TASK_TIMEZONE_SELECTOR_CUSTOM_ID,
					Placeholder: "Timezone",
					MinValues:   &one,
					MaxValues:   1,
					Options:     timezoneOpts,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    discord.STATS_TRACKER_TASK_WEEKDAYS_SELECTOR_CUSTOM_ID,
					Placeholder: "Weekdays",
					MinValues:   &one,
					MaxValues:   7,
					Options: []discordgo.SelectMenuOption{
						{
							Label:   p.Sprintf("Monday"),
							Value:   "1",
							Default: slices.Contains(s.LocalWeekdays, time.Monday),
						},
						{
							Label:   p.Sprintf("Tuesday"),
							Value:   "2",
							Default: slices.Contains(s.LocalWeekdays, time.Tuesday),
						},
						{
							Label:   p.Sprintf("Wednesday"),
							Value:   "3",
							Default: slices.Contains(s.LocalWeekdays, time.Wednesday),
						},
						{
							Label:   p.Sprintf("Thursday"),
							Value:   "4",
							Default: slices.Contains(s.LocalWeekdays, time.Thursday),
						},
						{
							Label:   p.Sprintf("Friday"),
							Value:   "5",
							Default: slices.Contains(s.LocalWeekdays, time.Friday),
						},
						{
							Label:   p.Sprintf("Saturday"),
							Value:   "6",
							Default: slices.Contains(s.LocalWeekdays, time.Saturday),
						},
						{
							Label:   p.Sprintf("Sunday"),
							Value:   "0",
							Default: slices.Contains(s.LocalWeekdays, time.Sunday),
						},
					},
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
					Options:     hourPickerOptions(p, selectedHour),
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
					Options:     minutePickerOptions(p, selectedMinute),
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
					Options:     durationPickerOptions(p, s.LocalEndTime-s.LocalStartTime),
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: discord.STATS_TRACKER_TASK_SUBMIT_BUTTON_CUSTOM_ID,
					Label:    p.Sprintf("Submit"),
					Style:    discordgo.SuccessButton,
				},
			},
		},
	}
}
