package character_loader

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type BatchLoader struct {
	cache       *expirable.LRU[string, ps2.Character]
	loader      loaders.QueriedLoader[[]string, map[string]ps2.Character]
	charId      chan string
	charIdBatch chan []string
	awaitersMu  sync.Mutex
	awaiters    map[string]chan ps2.Character
	batchRate   time.Duration
}

func NewBatch(
	loader loaders.QueriedLoader[[]string, map[string]ps2.Character],
	batchRate time.Duration,
) *BatchLoader {
	return &BatchLoader{
		cache:       expirable.NewLRU[string, ps2.Character](1000, nil, time.Hour*12),
		loader:      loader,
		charId:      make(chan string, 1000),
		charIdBatch: make(chan []string),
		awaiters:    make(map[string]chan ps2.Character),
		batchRate:   batchRate,
	}
}

func (l *BatchLoader) flushBatchTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.flushBatchTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	ticker := time.NewTicker(l.batchRate)
	defer ticker.Stop()
	var batch []string
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-l.charId:
			batch = append(batch, id)
		case <-ticker.C:
			log.Debug("flushing batch", slog.Int("batch_size", len(batch)))
			if len(batch) == 0 {
				continue
			}
			l.charIdBatch <- batch
			batch = make([]string, 0, len(batch))
		}
	}
}

func (l *BatchLoader) releaseAwaiters(log *slog.Logger, batch []string, chars map[string]ps2.Character) {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	log.Debug("releasing awaiters", slog.Int("batch_size", len(batch)))
	for _, id := range batch {
		if c, ok := l.awaiters[id]; ok {
			if char, ok := chars[id]; ok {
				c <- char
			}
			close(c)
			delete(l.awaiters, id)
		}
	}
}

func (l *BatchLoader) processBatchTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.processBatchTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case batch := <-l.charIdBatch:
			log.Debug("execute batch", slog.Int("batch_size", len(batch)))
			loaded, err := l.loader.Load(ctx, batch)
			if err != nil {
				log.Error("failed to load characters", sl.Err(err))
				loaded = make(map[string]ps2.Character)
			}
			l.releaseAwaiters(log, batch, loaded)
		}
	}
}

func (l *BatchLoader) cleanupTask(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.cleanupTask"
	log := infra.OpLogger(ctx, op)
	defer wg.Done()
	<-ctx.Done()
	log.Info("cleaning up")
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	for _, c := range l.awaiters {
		close(c)
	}
	clear(l.awaiters)
}

func (l *BatchLoader) Start(ctx context.Context, wg *sync.WaitGroup) {
	const op = "loaders.character_loader.batch.Start"
	infra.OpLogger(ctx, op).Info("starting loader")
	wg.Add(3)
	go l.flushBatchTask(ctx, wg)
	go l.processBatchTask(ctx, wg)
	go l.cleanupTask(ctx, wg)
}

func (l *BatchLoader) load(log *slog.Logger, charId string) chan ps2.Character {
	l.awaitersMu.Lock()
	defer l.awaitersMu.Unlock()
	log.Debug("awaiting for", slog.String("character_id", charId))
	c := make(chan ps2.Character)
	l.awaiters[charId] = c
	l.charId <- charId
	return c
}

func (l *BatchLoader) Load(ctx context.Context, charId string) (ps2.Character, error) {
	const op = "loaders.character_loader.batch.Load"
	log := infra.OpLogger(ctx, op)
	log.Debug("loading", slog.String("character_id", charId))
	cached, ok := l.cache.Get(charId)
	if ok {
		return cached, nil
	}
	char := <-l.load(log, charId)
	if char.Id == "" {
		return char, loaders.ErrNotFound
	}
	l.cache.Add(charId, char)
	return char, nil
}
