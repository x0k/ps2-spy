package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/shared"
	_ "github.com/x0k/ps2-spy/internal/translations"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func main() {
	tag := language.MustParse(string(discordgo.EnglishGB))
	p := message.NewPrinter(tag)
	m := discord_messages.New(
		shared.Timezones,
		4*time.Hour,
	)
	_, err := m.InvalidPopulationType("foo", nil)(p)
	println(err.Msg)
}
