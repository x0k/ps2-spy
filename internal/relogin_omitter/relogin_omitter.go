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
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrConvertEvent = fmt.Errorf("failed to convert event")

type ReLoginOmitter struct {
	publisher.Publisher[publisher.Event]
	platform          platforms.Platform
	log               *logger.Logger
	batchMu           sync.Mutex
	logoutEventsQueue *containers.ExpirationQueue[ps2.CharacterId]
	logoutEvents      map[ps2.CharacterId]ps2events.PlayerLogout
	flushInterval     time.Duration
	delayDuration     time.Duration
	mt                metrics.Metrics
}

func New(
	log *logger.Logger,
	platform platforms.Platform,
	pub publisher.Publisher[publisher.Event],
	mt metrics.Metrics,
) *ReLoginOmitter {
	return &ReLoginOmitter{
		Publisher:         pub,
		platform:          platform,
		log:               log.With(slog.String("component", "relogin_omitter.ReLoginOmitter")),
		logoutEventsQueue: containers.NewExpirationQueue[ps2.CharacterId](),
		logoutEvents:      make(map[ps2.CharacterId]ps2events.PlayerLogout),
		flushInterval:     1 * time.Minute,
		delayDuration:     3 * time.Minute,
		mt:                mt,
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
	r.mt.SetPlatformQueueSize(
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

func (r *ReLoginOmitter) flushTask(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
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

func (r *ReLoginOmitter) Start(ctx context.Context) {
	r.log.Info(ctx, "starting")
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
	return r.Publisher.Publish(event)
}
