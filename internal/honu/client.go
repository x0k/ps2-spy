package honu

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	httpClient   *http.Client
	honuEndpoint string
}

func NewClient(honuEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:   httpClient,
		honuEndpoint: honuEndpoint,
	}
}

func (c *Client) Endpoint() string { return c.honuEndpoint }

func (c *Client) WorldOverview(ctx context.Context) ([]World, error) {
	url := c.honuEndpoint + "/api/world/overview"
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
}
