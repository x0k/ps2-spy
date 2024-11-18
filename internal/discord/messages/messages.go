package discord_messages

import "github.com/x0k/ps2-spy/internal/ps2"

type Messages interface {
	CharacterLogin(ps2.Character) (string, error)
}
