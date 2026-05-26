package module

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"log/slog"
)

type Module struct {
	name     string
	log      *slog.Logger
	wg       sync.WaitGroup
	services []Runnable
	onStart  []Runnable
	onStop   []Runnable
	signal   bool
}

type Option func(*Module)

func WithSignalHandling() Option {
	return func(m *Module) {
		m.signal = true
	}
}

func New(log *slog.Logger, name string, opts ...Option) *Module {
	m := &Module{
		log:  log,
		name: name,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) awaiter(ctx context.Context) error {
	if m.signal {
		m.log.LogAttrs(ctx, slog.LevelInfo, "press CTRL-C to exit")
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(stop)
		select {
		case s := <-stop:
			m.log.LogAttrs(ctx, slog.LevelInfo, "received signal", slog.String("signal", s.String()))
			return nil
		case <-ctx.Done():
			return nil
		}
	}
	<-ctx.Done()
	return nil
}

func (m *Module) run(ctx context.Context) error {
	if len(m.services) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, hook := range m.onStart {
		m.log.LogAttrs(ctx, slog.LevelInfo, "run on start", slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			return err
		}
	}

	for _, service := range m.services {
		m.wg.Go(func() {
			m.log.LogAttrs(ctx, slog.LevelInfo, "starting", slog.String("service", service.Name()))
			if err := service.Run(ctx); err != nil {
				m.log.LogAttrs(ctx, slog.LevelError, "service failed", slog.String("service", service.Name()), slog.String("error", err.Error()))
			}
			m.log.LogAttrs(ctx, slog.LevelInfo, "stopped", slog.String("service", service.Name()))
		})
	}

	err := m.awaiter(ctx)

	m.log.LogAttrs(ctx, slog.LevelInfo, "stopping")
	cancel()

	for _, hook := range m.onStop {
		m.log.LogAttrs(ctx, slog.LevelInfo, "run on stop", slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			m.log.LogAttrs(ctx, slog.LevelError, "on stop hook failed", slog.String("hook", hook.Name()), slog.String("error", err.Error()))
		}
	}

	m.wg.Wait()

	return err
}

func (m *Module) Run(ctx context.Context) error {
	return m.run(ctx)
}
