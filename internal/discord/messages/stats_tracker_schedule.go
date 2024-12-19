package discord_messages

import (
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"golang.org/x/text/message"
)

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

func (m *Messages) statsTrackerScheduleEditForm(
	p *message.Printer,
	tasks []discord.StatsTrackerTask,
	offset time.Duration,
) []discordgo.MessageComponent {
	localTasks := make([]localStatsTrackerTask, 0, len(tasks))
	for _, t := range tasks {
		startWeekday, startTime := localDate(t.Weekday, t.UtcStartTime, offset)
		endWeekday, endTime := localDate(t.Weekday, t.UtcEndTime, offset)
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
				Label:    p.Sprintf("Add"),
				Style:    discordgo.PrimaryButton,
			},
		},
	})
	return rows
}
