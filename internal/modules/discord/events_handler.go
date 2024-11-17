package discord_module

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
)

type EventsHandler struct {
	session *discordgo.Session
}

func (h *EventsHandler) HandlePlayerLogin(e characters_tracker.PlayerLogin) {

}
