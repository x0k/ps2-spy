package relogin_omitter

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var ErrConvertEvent = fmt.Errorf("failed to convert event")

type ReLoginOmitter struct {
	pubsub.Publisher[events.Event]
	log               *logger.Logger
	batchMu           sync.Mutex
	logoutEventsQueue *containers.ExpirationQueue[ps2.CharacterId]
	logoutEvents      map[ps2.CharacterId]events.PlayerLogout
	flushInterval     time.Duration
	delayDuration     time.Duration
	// mt                metrics.Metrics
}

func NewReLoginOmitter(
	log *logger.Logger,
	pub pubsub.Publisher[events.Event],
	// mt metrics.Metrics,
) *ReLoginOmitter {
	return &ReLoginOmitter{
		Publisher:         pub,
		log:               log.With(sl.Component("platform.ReLoginOmitter")),
		logoutEventsQueue: containers.NewExpirationQueue[ps2.CharacterId](),
		logoutEvents:      make(map[ps2.CharacterId]events.PlayerLogout),
		flushInterval:     1 * time.Minute,
		delayDuration:     3 * time.Minute,
		// mt:                mt,
	}
}

func (r *ReLoginOmitter) addLogoutEvent(event *events.PlayerLogout) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	charId := ps2.CharacterId(event.CharacterID)
	r.logoutEventsQueue.Push(charId)
	r.logoutEvents[charId] = *event
}

func (r *ReLoginOmitter) shouldPublishLoginEvent(event *events.PlayerLogin) bool {
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

func (r *ReLoginOmitter) flushLogOutEvents(ctx context.Context, now time.Time) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	count := r.logoutEventsQueue.RemoveExpired(now.Add(-r.delayDuration), func(charId ps2.CharacterId) {
		if e, ok := r.logoutEvents[charId]; ok {
			if err := r.Publisher.Publish(&e); err != nil {
				r.log.Error(ctx, "failed to publish logout event", sl.Err(err))
			}
			delete(r.logoutEvents, charId)
		}
	})
	// r.mt.SetPlatformQueueSize(
	// 	metrics.LogoutEventsQueueName,
	// 	r.platform,
	// 	r.logoutEventsQueue.Len(),
	// )
	if count > 0 {
		r.log.Debug(
			ctx,
			"logout events flushed",
			slog.Int("events_count", count),
			slog.Int("queue_size", r.logoutEventsQueue.Len()),
		)
	}
}

func (r *ReLoginOmitter) Start(ctx context.Context) {
	ticker := time.NewTicker(r.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			r.flushLogOutEvents(ctx, now)
		}
	}
}

func (r *ReLoginOmitter) Publish(event pubsub.Event[events.EventType]) error {
	if event.Type() == events.PlayerLogoutEventName {
		if e, ok := event.(*events.PlayerLogout); ok {
			r.addLogoutEvent(e)
			return nil
		}
		return ErrConvertEvent
	}
	if event.Type() == events.PlayerLoginEventName {
		if e, ok := event.(*events.PlayerLogin); ok {
			if !r.shouldPublishLoginEvent(e) {
				return nil
			}
		} else {
			return ErrConvertEvent
		}
	}
	return r.Publisher.Publish(event)
}
