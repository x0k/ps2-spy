package voidwell

import (
	"context"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient       *http.Client
	voidwellEndpoint string
}

const worldsStateUrl = "/ps2/worldstate?platform=pc"

func NewClient(voidwellEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:       httpClient,
		voidwellEndpoint: voidwellEndpoint,
	}
}

func (c *Client) Endpoint() string { return c.voidwellEndpoint }

func (c *Client) WorldsState(ctx context.Context) ([]World, error) {
	url := c.voidwellEndpoint + worldsStateUrl
	return httpx.GetJson[[]World](ctx, c.httpClient, url)
}
