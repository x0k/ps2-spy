package main

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/facilities_manager"
	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
)

func startFacilitiesManager(
	ctx context.Context,
	ps2EventsPublisher *publisher.Publisher,
	facilitiesManager *facilities_manager.FacilitiesManager,
) error {
	const op = "startFacilitiesManager"
	log := infra.OpLogger(ctx, op)
	wg := infra.Wg(ctx)
	facilityControl := make(chan ps2events.FacilityControl)
	facilityControlUnSub, err := ps2EventsPublisher.AddHandler(facilityControl)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer facilityControlUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-facilityControl:
				err := facilitiesManager.FacilityControl(ctx, msg)
				if err != nil {
					log.Error("facility control", sl.Err(err))
				}
			}
		}
	}()
	return nil
}
