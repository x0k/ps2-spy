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
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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
	mt                *metrics.Metrics
	platform          ps2_platforms.Platform
	onError           func(err error)
}

func NewReLoginOmitter(
	log *logger.Logger,
	pub pubsub.Publisher[events.Event],
	mt *metrics.Metrics,
	platform ps2_platforms.Platform,
	onError func(err error),
) *ReLoginOmitter {
	return &ReLoginOmitter{
		Publisher:         pub,
		log:               log,
		logoutEventsQueue: containers.NewExpirationQueue[ps2.CharacterId](),
		logoutEvents:      make(map[ps2.CharacterId]events.PlayerLogout),
		flushInterval:     1 * time.Minute,
		delayDuration:     3 * time.Minute,
		mt:                mt,
		platform:          platform,
	}
}

func (r *ReLoginOmitter) addLogoutEvent(event events.PlayerLogout) {
	r.batchMu.Lock()
	defer r.batchMu.Unlock()
	charId := ps2.CharacterId(event.CharacterID)
	r.logoutEventsQueue.Push(charId)
	r.logoutEvents[charId] = event
}

func (r *ReLoginOmitter) shouldPublishLoginEvent(event events.PlayerLogin) bool {
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
			delete(r.logoutEvents, charId)
			r.Publisher.Publish(e)
		}
	})
	metrics.SetPlatformQueueSize(
		r.mt,
		metrics.LogoutEventsQueueName,
		r.platform,
		r.logoutEventsQueue.Len(),
	)
	if count > 0 {
		r.log.Debug(
			ctx,
			"logout events flushed",
			slog.Int("events_count", count),
			slog.Int("queue_size", r.logoutEventsQueue.Len()),
		)
	}
}

func (r *ReLoginOmitter) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			r.flushLogOutEvents(ctx, now)
		}
	}
}

func (r *ReLoginOmitter) Publish(event pubsub.Event[events.EventType]) {
	if event.Type() == events.PlayerLogoutEventName {
		if e, ok := event.(events.PlayerLogout); ok {
			r.addLogoutEvent(e)
			return
		}
		r.onError(ErrConvertEvent)
		return
	}
	if event.Type() == events.PlayerLoginEventName {
		if e, ok := event.(events.PlayerLogin); ok {
			if !r.shouldPublishLoginEvent(e) {
				return
			}
		} else {
			r.onError(ErrConvertEvent)
			return
		}
	}
	r.Publisher.Publish(event)
}
