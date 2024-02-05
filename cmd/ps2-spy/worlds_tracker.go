package main

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func startWorldsTracker(
	ctx context.Context,
	log *logger.Logger,
	worldsTracker *worlds_tracker.WorldsTracker,
	ps2EventsPublisher *ps2events.Publisher,
) {
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
					log.Error(ctx, "error handling metagame event", sl.Err(err))
				}
			case e := <-facilityControl:
				if err := worldsTracker.HandleFacilityControl(ctx, e); err != nil {
					log.Error(ctx, "error handling facility control", sl.Err(err))
				}
			case e := <-continentLock:
				if err := worldsTracker.HandleContinentLock(ctx, e); err != nil {
					log.Error(ctx, "error handling continent lock", sl.Err(err))
				}
			}
		}
	}()
}

func startNewWorldsTracker(
	ctx context.Context,
	log *logger.Logger,
	ps2EventsPublisher *ps2events.Publisher,
	worldsTrackerPublisher *worlds_tracker.Publisher,
) *worlds_tracker.WorldsTracker {
	worldsTracker := worlds_tracker.New(5*time.Minute, worldsTrackerPublisher)
	startWorldsTracker(ctx, log, worldsTracker, ps2EventsPublisher)
	return worldsTracker
}
