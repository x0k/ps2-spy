package module

import (
	"context"
	"sync"
	"sync/atomic"

	"log/slog"
)

type Module struct {
	name      string
	log       *slog.Logger
	wg        sync.WaitGroup
	services  []Runnable
	preStart  []Runnable
	postStart []Runnable
	preStop   []Runnable
	postStop  []Runnable
	fatal     chan error
	stopped   atomic.Bool
}

func New(log *slog.Logger, name string) *Module {
	return &Module{
		log:   log,
		name:  name,
		fatal: make(chan error, 1),
	}
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) awaiter(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-m.fatal:
		return err
	}
}

func (m *Module) run(ctx context.Context, awaiter func(context.Context) error) error {
	if len(m.services) == 0 && len(m.postStart) == 0 && len(m.preStop) == 0 {
		return nil
	}

	if m.stopped.Load() {
		return <-m.fatal
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, hook := range m.preStart {
		m.log.LogAttrs(ctx, slog.LevelInfo, "run pre start", slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			return err
		}
	}

	for _, service := range m.services {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.log.LogAttrs(ctx, slog.LevelInfo, "starting", slog.String("service", service.Name()))
			if err := service.Run(ctx); err != nil {
				m.Fatal(ctx, err)
			}
			m.log.LogAttrs(ctx, slog.LevelInfo, "stopped", slog.String("service", service.Name()))
		}()
	}

	for _, hook := range m.postStart {
		m.log.LogAttrs(ctx, slog.LevelInfo, "run post start", slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			m.Fatal(ctx, err)
		}
	}

	err := awaiter(ctx)

	for _, hook := range m.preStop {
		m.log.LogAttrs(ctx, slog.LevelInfo, "run pre stop", slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			m.Fatal(ctx, err)
		}
	}

	m.log.LogAttrs(ctx, slog.LevelInfo, "stopping")
	m.stopped.Store(true)
	cancel()

	for _, hook := range m.postStop {
		m.log.LogAttrs(ctx, slog.LevelInfo, "run post stop", slog.String("hook", hook.Name()))
		if err := hook.Run(ctx); err != nil {
			m.Fatal(ctx, err)
		}
	}

	m.wg.Wait()

	return err
}

func (m *Module) Run(ctx context.Context) error {
	return m.run(ctx, m.awaiter)
}
