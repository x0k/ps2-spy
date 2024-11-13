package events_module

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/relogin_omitter"
)

func NewReLoginOmitterService(
	platform platforms.Platform,
	reLoginOmitter *relogin_omitter.ReLoginOmitter,
) module.Service {
	return module.NewService(
		fmt.Sprintf("platform.%s.relogin_omitter", platform),
		func(ctx context.Context) error {
			reLoginOmitter.Start(ctx)
			return nil
		},
	)
}
