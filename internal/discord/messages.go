package discord

import "github.com/x0k/ps2-spy/internal/ps2"

type Error[M any] struct {
	Msg M
	Err error
}
type Localized = func(locale Locale) string

type LocalizedError = Error[Localized]

type StringError = Error[string]

type Messages interface {
	About() string
	CharacterLogin(ps2.Character) (string, *StringError)
	CharacterLoadError(ps2.CharacterId, error) (string, *StringError)
}

type LocalizedMessage = func(locale Locale) (string, *StringError)

type LocalizedMessages interface {
	About() Localized
	CharacterLogin(ps2.Character) LocalizedMessage
	CharacterLoadError(ps2.CharacterId, error) LocalizedMessage
}
