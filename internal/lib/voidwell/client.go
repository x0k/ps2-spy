package voidwell

import (
	"context"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient       *http.Client
	voidwellEndpoint string
	worlds           *containers.ExpiableValue[[]World]
}

const worldsStateUrl = "/ps2/worldstate?platform=pc"

func NewClient(voidwellEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:       httpClient,
		voidwellEndpoint: voidwellEndpoint,
		worlds:           containers.NewExpiableValue[[]World](time.Minute),
	}
}

func (c *Client) Start() {
	go c.worlds.StartExpiration()
}

func (c *Client) Stop() {
	c.worlds.StopExpiration()
}

func (c *Client) Endpoint() string { return c.voidwellEndpoint }

func (c *Client) WorldsState(ctx context.Context) ([]World, error) {
	return c.worlds.Load(func() ([]World, error) {
		url := c.voidwellEndpoint + worldsStateUrl
		return httpx.GetJson[[]World](ctx, c.httpClient, url)
	})
}
