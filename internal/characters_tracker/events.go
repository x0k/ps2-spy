package characters_tracker

import (
	"time"

	"github.com/x0k/ps2-spy/internal/ps2"
)

const (
	PlayerLoginType  = "player_login"
	PlayerLogoutType = "player_logout"
)

type PlayerLogin struct {
	Time        time.Time
	CharacterId ps2.CharacterId
	WorldId     ps2.WorldId
}

func (e PlayerLogin) Type() string {
	return PlayerLoginType
}

type PlayerLogout struct {
	Time        time.Time
	CharacterId ps2.CharacterId
	WorldId     ps2.WorldId
}

func (e PlayerLogout) Type() string {
	return PlayerLogoutType
}
