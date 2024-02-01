package main

import (
	"context"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func startWorldsTracker(
	ctx context.Context,
	log *logger.Logger,
	worldsTracker *worlds_tracker.WorldsTracker,
	ps2EventsPublisher *publisher.Publisher,
) error {
	const op = "startWorldsTracker"
	worldsTracker.Start(ctx)
	metagameEvent := make(chan ps2events.MetagameEvent)
	metagameEventUnSub, err := ps2EventsPublisher.AddHandler(metagameEvent)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	facilityControl := make(chan ps2events.FacilityControl)
	facilityControlUnSub, err := ps2EventsPublisher.AddHandler(facilityControl)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	continentLock := make(chan ps2events.ContinentLock)
	continentLockUnSub, err := ps2EventsPublisher.AddHandler(continentLock)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
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
	return nil
}

func startNewWorldsTracker(
	ctx context.Context,
	log *logger.Logger,
	ps2EventsPublisher *publisher.Publisher,
	worldsTrackerPublisher publisher.Abstract[publisher.Event],
) (*worlds_tracker.WorldsTracker, error) {
	worldsTracker := worlds_tracker.New(5*time.Minute, worldsTrackerPublisher)
	return worldsTracker, startWorldsTracker(ctx, log, worldsTracker, ps2EventsPublisher)
}
