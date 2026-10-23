package module_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/module"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestModuleStopsOnServiceError(t *testing.T) {
	m := module.New(testLogger(), "test")
	expectedErr := errors.New("boom")
	m.Go(module.NewRun("failing", func(context.Context) error {
		return expectedErr
	}))
	m.Go(module.NewRun("blocking", func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}))

	result := make(chan error, 1)
	go func() {
		result <- m.Run(context.Background())
	}()

	select {
	case err := <-result:
		if !errors.Is(err, expectedErr) {
			t.Fatalf("Run() error = %v, want %v", err, expectedErr)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after service error")
	}
}

func TestModuleKeepsRunningOnFeatureError(t *testing.T) {
	m := module.New(testLogger(), "test")
	failed := make(chan struct{})
	m.GoFeature(module.NewRun("failing", func(context.Context) error {
		close(failed)
		return errors.New("boom")
	}))
	started := make(chan struct{})
	m.Go(module.NewRun("blocking", func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		return nil
	}))

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- m.Run(ctx)
	}()

	<-started
	<-failed

	select {
	case err := <-result:
		t.Fatalf("Run() stopped after feature error: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after context cancellation")
	}
}

func TestModuleIgnoresShutdownErrors(t *testing.T) {
	m := module.New(testLogger(), "test")
	started := make(chan struct{})
	m.Go(module.NewRun("cancelled", func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}))

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- m.Run(ctx)
	}()

	<-started
	cancel()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after context cancellation")
	}
}
