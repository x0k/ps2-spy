package storage

import (
	"fmt"
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/publisher"
)

type TxPublisher struct {
	pub    publisher.Abstract[publisher.Event]
	buffer []publisher.Event
}

func NewTxPublisher(pub publisher.Abstract[publisher.Event], estimatedEventsCount int) *TxPublisher {
	return &TxPublisher{
		pub:    pub,
		buffer: make([]publisher.Event, 0, estimatedEventsCount),
	}
}

func (b *TxPublisher) Publish(event publisher.Event) error {
	b.buffer = append(b.buffer, event)
	return nil
}

func (b *TxPublisher) Commit() error {
	errors := make([]string, 0, len(b.buffer))
	for _, event := range b.buffer {
		err := b.pub.Publish(event)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	b.buffer = nil
	if len(errors) > 0 {
		return fmt.Errorf("failed to publish events: %s", strings.Join(errors, ", "))
	}
	return nil
}
