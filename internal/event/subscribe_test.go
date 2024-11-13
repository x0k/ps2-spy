package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/x0k/ps2-spy/internal/event"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type TestType event.Type

const (
	FooTestType TestType = "foo"
	BarTestType TestType = "bar"
)

type FooEvent struct {
	Foo string
}

func (FooEvent) Type() event.Type {
	return event.Type(FooTestType)
}

type BarEvent struct {
	Bar string
}

func (BarEvent) Type() event.Type {
	return event.Type(BarTestType)
}

func TestSubscribe(t *testing.T) {
	ctx := context.Background()
	m := module.NewTestModule(t)
	defer m.RunPreStop(ctx)

	testPubSub := pubsub.New[event.Type]()
	foo := event.Subscribe[FooEvent](m, testPubSub)
	bar := event.Subscribe[BarEvent](m, testPubSub)

	go func() {
		if err := testPubSub.Publish(FooEvent{}); err != nil {
			return
		}
		if err := testPubSub.Publish(BarEvent{}); err != nil {
			return
		}
		if err := testPubSub.Publish(FooEvent{}); err != nil {
			return
		}
	}()

	fooCount := 0
	barCount := 0
	for fooCount+barCount < 3 {
		select {
		case <-foo:
			fooCount++
		case <-bar:
			barCount++
		case <-time.After(time.Second):
			t.Fatal("timed out")
		}
	}
	if fooCount != 2 {
		t.Errorf("expected 1 foo event, got %d", fooCount)
	}
	if barCount != 1 {
		t.Errorf("expected 1 bar event, got %d", barCount)
	}
}
