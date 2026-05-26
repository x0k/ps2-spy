package http_adapters

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/module"
)

func NewService(name string, srv *http.Server, log *slog.Logger) module.Runnable {
	return module.NewRun(name, func(ctx context.Context) error {
		context.AfterFunc(ctx, func() {
			if err := srv.Shutdown(context.Background()); err != nil {
				log.LogAttrs(ctx, slog.LevelError, "failed to shutdown server", slog.String("error", err.Error()))
			}
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
}
