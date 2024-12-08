package discord_event_handlers

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
)

type HandlersManager struct {
	log     *logger.Logger
	session *discordgo.Session

	wg              sync.WaitGroup
	handlersTimeout time.Duration

	mu       sync.Mutex
	cancelId int64
	cancels  map[int64]func()
}

func NewHandlersManager(
	log *logger.Logger,
	session *discordgo.Session,
	handlersTimeout time.Duration,
) *HandlersManager {
	return &HandlersManager{
		log:             log,
		session:         session,
		handlersTimeout: handlersTimeout,
		cancels:         make(map[int64]func()),
	}
}

func (h *HandlersManager) Start(ctx context.Context) {
	<-ctx.Done()
	h.mu.Lock()
	for _, sub := range h.cancels {
		sub()
	}
	clear(h.cancels)
	h.mu.Unlock()
	h.wg.Wait()
}

func (h *HandlersManager) addCancel(action func()) func() {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.cancelId
	h.cancelId++
	h.cancels[id] = action
	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.cancels, id)
	}
}
