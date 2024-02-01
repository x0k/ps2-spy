package characters_tracker

import "github.com/x0k/ps2-spy/internal/lib/publisher"

type playerLoginHandler chan<- PlayerLogin

func (h playerLoginHandler) Type() string {
	return PlayerLoginType
}

func (h playerLoginHandler) Handle(e publisher.Event) {
	h <- e.(PlayerLogin)
}

type playerLogoutHandler chan<- PlayerLogout

func (h playerLogoutHandler) Type() string {
	return PlayerLogoutType
}

func (h playerLogoutHandler) Handle(e publisher.Event) {
	h <- e.(PlayerLogout)
}

func CastHandler(handler any) publisher.Handler {
	switch v := handler.(type) {
	case chan PlayerLogin:
		return playerLoginHandler(v)
	case chan<- PlayerLogin:
		return playerLoginHandler(v)
	case chan PlayerLogout:
		return playerLogoutHandler(v)
	case chan<- PlayerLogout:
		return playerLogoutHandler(v)
	}
	return nil
}
