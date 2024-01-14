package login

import (
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
)

func New() handlers.Ps2EventHandler[ps2events.PlayerLogin] {
	return
}
