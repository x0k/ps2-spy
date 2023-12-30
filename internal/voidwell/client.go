package voidwell

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
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
