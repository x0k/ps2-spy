package discord

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

type ChannelTitleUpdater struct {
	log     *logger.Logger
	session *discordgo.Session
	mu      sync.Mutex
	titles  map[ChannelId]string
}

func NewChannelTitleUpdater(log *logger.Logger, session *discordgo.Session) *ChannelTitleUpdater {
	return &ChannelTitleUpdater{
		log:     log,
		session: session,
		titles:  make(map[ChannelId]string),
	}
}

func (c *ChannelTitleUpdater) Start(ctx context.Context) {
	ticker := time.NewTicker((10 * time.Minute) + (5 * time.Second))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.update(ctx)
		}
	}
}

func (c *ChannelTitleUpdater) UpdateTitle(channelId ChannelId, title string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.titles[channelId] = title
}

func (c *ChannelTitleUpdater) takeTitles() map[ChannelId]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	titles := c.titles
	c.titles = make(map[ChannelId]string, len(titles))
	return titles
}

func (c *ChannelTitleUpdater) update(ctx context.Context) {
	titles := c.takeTitles()
	println("updating channel titles", len(titles))
	for channelId, title := range titles {
		if _, err := c.session.ChannelEdit(string(channelId), &discordgo.ChannelEdit{
			Name: title,
		}); err != nil {
			c.log.Error(ctx, "failed to update channel title", slog.String("channel_id", string(channelId)), sl.Err(err))
		}
	}
}
