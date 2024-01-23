package relogin_event_omitter

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
)

var ErrConvertEvent = fmt.Errorf("failed to convert event")

type ReLoginOmitter struct {
	pub               publisher.Abstract[publisher.Event]
	batchMu           sync.Mutex
	logoutEventsBatch map[string]*ps2events.PlayerLogout
	batchInterval     time.Duration
}

func New(pub publisher.Abstract[publisher.Event]) *ReLoginOmitter {
	return &ReLoginOmitter{
		pub:               pub,
		logoutEventsBatch: make(map[string]*ps2events.PlayerLogout, 100),
		batchInterval:     time.Minute * 3,
	}
}

func (r *ReLoginOmitter) addLogoutEvent(event *ps2events.PlayerLogout) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	r.logoutEventsBatch[event.CharacterID] = event
}

func (r *ReLoginOmitter) shouldPublishLoginEvent(event *ps2events.PlayerLogin) bool {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	if _, ok := r.logoutEventsBatch[event.CharacterID]; ok {
		delete(r.logoutEventsBatch, event.CharacterID)
		return false
	}
	return true
}

func (r *ReLoginOmitter) flushLogOutEvents(log *slog.Logger) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	log.Debug("flush logout events", slog.Any("events_count", len(r.logoutEventsBatch)))
	for _, event := range r.logoutEventsBatch {
		r.pub.Publish(event)
	}
	r.logoutEventsBatch = make(map[string]*ps2events.PlayerLogout, len(r.logoutEventsBatch))
}

func (r *ReLoginOmitter) flushTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "relogin_event_ommiter.ReLoginOmitter.flushTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	ticker := time.NewTicker(r.batchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.flushLogOutEvents(log)
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
