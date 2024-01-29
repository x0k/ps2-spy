package relogin_omitter

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var ErrConvertEvent = fmt.Errorf("failed to convert event")

type ReLoginOmitter struct {
	pub               publisher.Abstract[publisher.Event]
	batchMu           sync.Mutex
	logoutEventsQueue *containers.ExpirationQueue[ps2.CharacterId]
	logoutEvents      map[ps2.CharacterId]ps2events.PlayerLogout
	flushInterval     time.Duration
	delayDuration     time.Duration
}

func New(pub publisher.Abstract[publisher.Event]) *ReLoginOmitter {
	return &ReLoginOmitter{
		pub:               pub,
		logoutEventsQueue: containers.NewExpirationQueue[ps2.CharacterId](),
		logoutEvents:      make(map[ps2.CharacterId]ps2events.PlayerLogout),
		flushInterval:     1 * time.Minute,
		delayDuration:     3 * time.Minute,
	}
}

func (r *ReLoginOmitter) addLogoutEvent(event *ps2events.PlayerLogout) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	charId := ps2.CharacterId(event.CharacterID)
	r.logoutEventsQueue.Push(charId)
	r.logoutEvents[charId] = *event
}

func (r *ReLoginOmitter) shouldPublishLoginEvent(event *ps2events.PlayerLogin) bool {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	charId := ps2.CharacterId(event.CharacterID)
	if r.logoutEventsQueue.Has(charId) {
		r.logoutEventsQueue.Remove(charId)
		delete(r.logoutEvents, charId)
		return false
	}
	return true
}

func (r *ReLoginOmitter) flushLogOutEvents(log *slog.Logger, now time.Time) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	count := r.logoutEventsQueue.RemoveExpired(now.Add(-r.delayDuration), func(charId ps2.CharacterId) {
		if e, ok := r.logoutEvents[charId]; ok {
			r.pub.Publish(&e)
			delete(r.logoutEvents, charId)
		}
	})
	if count > 0 {
		log.Debug(
			"logout events flushed",
			slog.Int("events_count", count),
			slog.Int("queue_size", r.logoutEventsQueue.Len()),
		)
	}
}

func (r *ReLoginOmitter) flushTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "relogin_event_ommiter.ReLoginOmitter.flushTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	ticker := time.NewTicker(r.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			r.flushLogOutEvents(log, now)
		}
	}
}

func (r *ReLoginOmitter) Start(ctx context.Context) {
	const op = "relogin_event_ommiter.ReLoginOmitter.Start"
	infra.OpLogger(ctx, op).Info("starting")
	wg := infra.Wg(ctx)
	wg.Add(1)
	go r.flushTask(ctx, wg)
}

func (r *ReLoginOmitter) Publish(event publisher.Event) error {
	if event.Type() == ps2events.PlayerLogoutEventName {
		if e, ok := event.(*ps2events.PlayerLogout); ok {
			r.addLogoutEvent(e)
			return nil
		}
		return ErrConvertEvent
	}
	if event.Type() == ps2events.PlayerLoginEventName {
		if e, ok := event.(*ps2events.PlayerLogin); ok {
			if !r.shouldPublishLoginEvent(e) {
				return nil
			}
		} else {
			return ErrConvertEvent
		}
	}
	return r.pub.Publish(event)
}
