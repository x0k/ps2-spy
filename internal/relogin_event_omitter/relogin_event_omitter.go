package relogin_event_omitter

import (
	"context"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/publisher"
)

type ReLoginOmitter struct {
	pub               publisher.Abstract[map[string]any]
	batchMu           sync.Mutex
	logoutEventsBatch map[string]map[string]any
	batchInterval     time.Duration
}

func New(pub publisher.Abstract[map[string]any]) *ReLoginOmitter {
	return &ReLoginOmitter{
		pub:               pub,
		logoutEventsBatch: make(map[string]map[string]any, 100),
		batchInterval:     time.Minute * 3,
	}
}

func isLoginEvent(event map[string]any) bool {
	return event[core.EventNameField].(string) == ps2events.PlayerLoginEventName
}

func isLogoutEvent(event map[string]any) bool {
	return event[core.EventNameField].(string) == ps2events.PlayerLogoutEventName
}

func characterId(event map[string]any) string {
	return event[ps2events.CharacterIdField].(string)
}

func (r *ReLoginOmitter) addLogoutEvent(event map[string]any) {
	characterId := characterId(event)
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	r.logoutEventsBatch[characterId] = event
}

func (r *ReLoginOmitter) shouldPublishLoginEvent(event map[string]any) bool {
	characterId := characterId(event)
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	if _, ok := r.logoutEventsBatch[characterId]; ok {
		delete(r.logoutEventsBatch, characterId)
		return false
	}
	return true
}

func (r *ReLoginOmitter) flushLogOutEvents() {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	for _, event := range r.logoutEventsBatch {
		r.pub.Publish(event)
	}
	r.logoutEventsBatch = make(map[string]map[string]any, len(r.logoutEventsBatch))
}

func (r *ReLoginOmitter) worker(ctx context.Context, wg *sync.WaitGroup) {
	const op = "relogin_event_ommiter.ReLoginOmitter.worker"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	ticker := time.NewTicker(r.batchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Debug("flush logout events")
			r.flushLogOutEvents()
		}
	}
}

func (r *ReLoginOmitter) Start(ctx context.Context) {
	const op = "relogin_event_ommiter.ReLoginOmitter.Start"
	infra.OpLogger(ctx, op).Info("starting")
	wg := infra.Wg(ctx)
	wg.Add(1)
	go r.worker(ctx, wg)
}

func (r *ReLoginOmitter) Publish(event map[string]any) error {
	if isLogoutEvent(event) {
		r.addLogoutEvent(event)
		return nil
	}
	if isLoginEvent(event) && !r.shouldPublishLoginEvent(event) {
		return nil
	}
	return r.pub.Publish(event)
}
