package http_adapters

import (
	"context"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/module"
)

func NewService(name string, srv *http.Server, fataler module.Fataler) module.Runnable {
	return module.NewRun(name, func(ctx context.Context) error {
		context.AfterFunc(ctx, func() {
			if err := srv.Shutdown(ctx); err != nil {
				fataler.Fatal(ctx, err)
			}
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
}
