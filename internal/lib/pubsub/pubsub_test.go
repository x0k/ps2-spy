package pubsub_test

import (
	"testing"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type eventType string

const testEventType eventType = "test"

type testEvent struct{}

func (testEvent) Type() eventType {
	return testEventType
}

type blockingHandler struct {
	started chan struct{}
	release chan struct{}
}

func (h *blockingHandler) Type() eventType {
	return testEventType
}

func (h *blockingHandler) Handle(pubsub.Event[eventType]) {
	close(h.started)
	<-h.release
}

func TestPublishDoesNotHoldLockWhileCallingHandlers(t *testing.T) {
	p := pubsub.New[eventType]()
	h := &blockingHandler{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	unsubscribe := p.AddHandler(h)

	published := make(chan struct{})
	go func() {
		p.Publish(testEvent{})
		close(published)
	}()

	<-h.started

	unsubscribed := make(chan struct{})
	go func() {
		unsubscribe()
		close(unsubscribed)
	}()

	select {
	case <-unsubscribed:
	case <-time.After(time.Second):
		t.Fatal("unsubscribe blocked while handler was running")
	}

	close(h.release)
	<-published
}
