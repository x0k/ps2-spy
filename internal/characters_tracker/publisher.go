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

type Publisher struct {
	publisher.Publisher[publisher.Event]
}

func NewPublisher(pub publisher.Publisher[publisher.Event]) *Publisher {
	return &Publisher{
		Publisher: pub,
	}
}

func (p *Publisher) AddPlayerLoginHandler(c chan<- PlayerLogin) func() {
	return p.AddHandler(playerLoginHandler(c))
}

func (p *Publisher) AddPlayerLogoutHandler(c chan<- PlayerLogout) func() {
	return p.AddHandler(playerLogoutHandler(c))
}
