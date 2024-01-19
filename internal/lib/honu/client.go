package honu

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient   *http.Client
	honuEndpoint string
	worlds       *containers.ExpiableValue[[]World]
}

const worldOverviewUrl = "/api/world/overview"

func NewClient(honuEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:   httpClient,
		honuEndpoint: honuEndpoint,
		worlds:       containers.NewExpiableValue[[]World](time.Minute),
	}
}

func (c *Client) Start(ctx context.Context, wg *sync.WaitGroup) {
	c.worlds.Start(ctx, wg)
}

func (c *Client) Endpoint() string { return c.honuEndpoint }

func (c *Client) WorldOverview(ctx context.Context) ([]World, error) {
	return c.worlds.Load(func() ([]World, error) {
		url := c.honuEndpoint + worldOverviewUrl
		return httpx.GetJson[[]World](ctx, c.httpClient, url)
	})
}
