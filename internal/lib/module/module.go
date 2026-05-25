package module

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
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
	fatal    chan error
	stopped  atomic.Bool
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
		log:   log,
		name:  name,
		fatal: make(chan error, 1),
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
		case err := <-m.fatal:
			return err
		}
	}
	select {
	case <-ctx.Done():
		return nil
	case err := <-m.fatal:
		return err
	}
}

func (m *Module) run(ctx context.Context) error {
	if len(m.services) == 0 {
		return nil
	}

	if m.stopped.Load() {
		return <-m.fatal
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
				m.Fatal(ctx, err)
			}
			m.log.LogAttrs(ctx, slog.LevelInfo, "stopped", slog.String("service", service.Name()))
		})
	}

	err := m.awaiter(ctx)

	m.log.LogAttrs(ctx, slog.LevelInfo, "stopping")
	m.stopped.Store(true)
	cancel()

	for _, hook := range m.onStop {
		m.log.LogAttrs(ctx, slog.LevelInfo, "run on stop", slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			m.Fatal(ctx, err)
		}
	}

	m.wg.Wait()

	return err
}

func (m *Module) Run(ctx context.Context) error {
	return m.run(ctx)
}
