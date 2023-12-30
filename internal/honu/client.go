package honu

import (
	"context"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
	"github.com/x0k/ps2-spy/internal/httpx"
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

func (c *Client) Start() {
	go c.worlds.StartExpiration()
}

func (c *Client) Stop() {
	c.worlds.StopExpiration()
}

func (c *Client) Endpoint() string { return c.honuEndpoint }

func (c *Client) WorldOverview(ctx context.Context) ([]World, error) {
	return c.worlds.Load(func() ([]World, error) {
		url := c.honuEndpoint + worldOverviewUrl
		var contentBody []World
		err := httpx.GetJson(ctx, c.httpClient, url, &contentBody)
		if err != nil {
			return nil, err
		}
		return contentBody, nil
	})
}
