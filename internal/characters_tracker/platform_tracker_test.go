package characters_tracker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type characterEventPublisher struct {
	events []Event
}

func (p *characterEventPublisher) Publish(event pubsub.Event[EventType]) {
	p.events = append(p.events, event)
}

func TestPlatformTrackerPublishesPlatformLoginAndLogout(t *testing.T) {
	publisher := &characterEventPublisher{}
	tracker := newCharactersTracker(
		logger.New(slog.New(slog.NewTextHandler(io.Discard, nil))),
		ps2_platforms.PC,
		[]ps2.WorldId{"17"},
		func(ctx context.Context, platform ps2_platforms.Platform, id ps2.CharacterId) (ps2.Character, error) {
			return ps2.Character{
				Id:       id,
				WorldId:  "17",
				Platform: platform,
			}, nil
		},
		publisher,
		nil,
	)

	tracker.HandleLogin(context.Background(), events.PlayerLogin{
		EventBase:   core.EventBase{Timestamp: "10"},
		CharacterID: "char-1",
	})
	tracker.wg.Wait()
	tracker.HandleLogout(context.Background(), events.PlayerLogout{
		EventBase:   core.EventBase{Timestamp: "20"},
		CharacterID: "char-1",
		WorldID:     "17",
	})

	if len(publisher.events) != 2 {
		t.Fatalf("events = %#v", publisher.events)
	}
	login, ok := publisher.events[0].(PlayerLogin)
	if !ok {
		t.Fatalf("event[0] = %T, want PlayerLogin", publisher.events[0])
	}
	if login.Platform != ps2_platforms.PC || login.Character.Id != "char-1" || !login.Time.Equal(time.Unix(10, 0)) {
		t.Fatalf("login = %#v", login)
	}
	logout, ok := publisher.events[1].(PlayerLogout)
	if !ok {
		t.Fatalf("event[1] = %T, want PlayerLogout", publisher.events[1])
	}
	if logout.Platform != ps2_platforms.PC || logout.CharacterId != "char-1" || !logout.Time.Equal(time.Unix(20, 0)) {
		t.Fatalf("logout = %#v", logout)
	}
}
