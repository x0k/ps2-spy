package main

import (
	"context"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func startWorldsTracker(
	ctx context.Context,
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
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer metagameEventUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-metagameEvent:
				worldsTracker.HandleMetagameEvent(ctx, e)
			}
		}
	}()
	return nil
}

func startNewWorldsTracker(
	ctx context.Context,
	ps2EventsPublisher *publisher.Publisher,
) (*worlds_tracker.WorldsTracker, error) {
	const op = "startNewWorldsTracker"
	log := infra.OpLogger(ctx, op)
	worldsTracker := worlds_tracker.New(log, 5*time.Minute)
	return worldsTracker, startWorldsTracker(ctx, worldsTracker, ps2EventsPublisher)
}
