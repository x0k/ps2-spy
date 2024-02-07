package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func startNewWorldsTracker(
	ctx context.Context,
	log *logger.Logger,
	platform platforms.Platform,
	ps2EventsPublisher *ps2events.Publisher,
	worldsTrackerPublisher *worlds_tracker.Publisher,
	platformWorldMapLoader loaders.KeyedLoader[ps2.WorldId, ps2.WorldMap],
) *worlds_tracker.WorldsTracker {
	pLog := log.With(slog.String("platform", string(platform)))
	worldsTracker := worlds_tracker.New(
		pLog,
		platform,
		5*time.Minute,
		worldsTrackerPublisher,
		platformWorldMapLoader,
	)
	worldsTracker.Start(ctx)
	metagameEvent := make(chan ps2events.MetagameEvent)
	metagameEventUnSub := ps2EventsPublisher.AddMetagameEventHandler(metagameEvent)
	facilityControl := make(chan ps2events.FacilityControl)
	facilityControlUnSub := ps2EventsPublisher.AddFacilityControlHandler(facilityControl)
	continentLock := make(chan ps2events.ContinentLock)
	continentLockUnSub := ps2EventsPublisher.AddContinentLockHandler(continentLock)
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer metagameEventUnSub()
		defer facilityControlUnSub()
		defer continentLockUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-metagameEvent:
				if err := worldsTracker.HandleMetagameEvent(ctx, e); err != nil {
					pLog.Error(ctx, "error handling metagame event", sl.Err(err))
				}
			case e := <-facilityControl:
				if err := worldsTracker.HandleFacilityControl(ctx, e); err != nil {
					pLog.Error(ctx, "error handling facility control", sl.Err(err))
				}
			case e := <-continentLock:
				if err := worldsTracker.HandleContinentLock(ctx, e); err != nil {
					pLog.Error(ctx, "error handling continent lock", sl.Err(err))
				}
			}
		}
	}()
	return worldsTracker
}
