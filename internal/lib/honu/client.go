package honu

import (
	"context"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient   *http.Client
	honuEndpoint string
}

const worldOverviewUrl = "/api/world/overview"

func NewClient(honuEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:   httpClient,
		honuEndpoint: honuEndpoint,
	}
}

func (c *Client) Endpoint() string { return c.honuEndpoint }

func (c *Client) WorldOverview(ctx context.Context) ([]World, error) {
	url := c.honuEndpoint + worldOverviewUrl
	return httpx.GetJson[[]World](ctx, c.httpClient, url)
}
