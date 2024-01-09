package main

import (
	"context"
	"log/slog"
	"sync"
)

type Setup struct {
	log *slog.Logger
	ctx context.Context
	wg  *sync.WaitGroup
}
