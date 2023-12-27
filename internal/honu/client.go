package honu

import (
	"encoding/json"
	"io"
	"net/http"
)

type client struct {
	httpClient   *http.Client
	honuEndpoint string
}

func NewClient(honuEndpoint string, httpClient *http.Client) *client {
	return &client{
		httpClient:   httpClient,
		honuEndpoint: honuEndpoint,
	}
}

func (c *client) WorldOverview() ([]World, error) {
	url := c.honuEndpoint + "/api/world/overview"
	resp, err := c.httpClient.Get(url)
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
