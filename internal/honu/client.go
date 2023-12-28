package honu

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/x0k/ps2-feed/internal/cache"
)

type Client struct {
	httpClient   *http.Client
	honuEndpoint string
	worlds       *cache.ExpiableValue[[]World]
}

const worldOverviewUrl = "/api/world/overview"

func NewClient(honuEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:   httpClient,
		honuEndpoint: honuEndpoint,
		worlds:       cache.NewExpiableValue[[]World](time.Minute),
	}
}

func (c *Client) Stop() { c.worlds.Stop() }

func (c *Client) Endpoint() string { return c.honuEndpoint }

func (c *Client) WorldOverview(ctx context.Context) ([]World, error) {
	return c.worlds.Load(func() ([]World, error) {
		url := c.honuEndpoint + worldOverviewUrl
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var contentBody []World
		err = json.Unmarshal(body, &contentBody)
		if err != nil {
			return nil, err
		}
		return contentBody, nil
	})
}
