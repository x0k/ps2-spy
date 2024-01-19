package relogin_event_omitter

import (
	"context"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
)

type AbstractPublisher interface {
	Publish(event map[string]any) error
}

type ReLoginOmitter struct {
	publisher         AbstractPublisher
	batchMu           sync.Mutex
	logoutEventsBatch map[string]map[string]any
	batchInterval     time.Duration
}

func New(publisher AbstractPublisher) *ReLoginOmitter {
	return &ReLoginOmitter{
		publisher:         publisher,
		logoutEventsBatch: make(map[string]map[string]any, 100),
		batchInterval:     time.Minute * 5,
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

func (r *ReLoginOmitter) publishLogOutEvents() {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	for _, event := range r.logoutEventsBatch {
		r.publisher.Publish(event)
	}
	r.logoutEventsBatch = make(map[string]map[string]any, len(r.logoutEventsBatch))
}

func (r *ReLoginOmitter) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(r.batchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.publishLogOutEvents()
		}
	}
}

func (r *ReLoginOmitter) Start(ctx context.Context) {
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
	return r.publisher.Publish(event)
}
