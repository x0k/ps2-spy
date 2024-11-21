package module

import (
	"context"
	"testing"
)

func NewTestModule(t *testing.T) *testModule {
	return &testModule{
		t: t,
	}
}

type testModule struct {
	t       *testing.T
	preStop []Hook
}

func (s *testModule) PreStop(hooks ...Hook) {
	s.preStop = append(s.preStop, hooks...)
}

func (s *testModule) RunPreStop(ctx context.Context) {
	for _, hook := range s.preStop {
		if err := hook.Run(ctx); err != nil {
			s.t.Fatal(err)
		}
	}
}
