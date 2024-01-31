package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func startWorldsTracker(
	ctx context.Context,
	worldsTracker *worlds_tracker.WorldsTracker,
	ps2EventsPublisher *publisher.Publisher,
) error {
	const op = "startWorldsTracker"
	log := infra.OpLogger(ctx, op)
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
					log.LogAttrs(ctx, slog.LevelError, "error handling metagame event", sl.Err(err))
				}
			case e := <-facilityControl:
				if err := worldsTracker.HandleFacilityControl(ctx, e); err != nil {
					log.LogAttrs(ctx, slog.LevelError, "error handling facility control", sl.Err(err))
				}
			case e := <-continentLock:
				if err := worldsTracker.HandleContinentLock(ctx, e); err != nil {
					log.LogAttrs(ctx, slog.LevelError, "error handling continent lock", sl.Err(err))
				}
			}
		}
	}()
	return nil
}

func startNewWorldsTracker(
	ctx context.Context,
	ps2EventsPublisher *publisher.Publisher,
	worldsTrackerPublisher publisher.Abstract[publisher.Event],
) (*worlds_tracker.WorldsTracker, error) {
	const op = "startNewWorldsTracker"
	log := infra.OpLogger(ctx, op)
	worldsTracker := worlds_tracker.New(log, 5*time.Minute, worldsTrackerPublisher)
	return worldsTracker, startWorldsTracker(ctx, worldsTracker, ps2EventsPublisher)
}
