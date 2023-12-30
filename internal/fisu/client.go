package fisu

import (
	"context"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
	"github.com/x0k/ps2-spy/internal/httpx"
)

type Client struct {
	httpClient       *http.Client
	fisuEndpoint     string
	worldsPopulation *containers.ExpiableValue[WorldsPopulation]
}

const populationApiUrl = "/api/population/?world=1,10,13,17,19,24,40,1000,2000"

func NewClient(fisuEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:       httpClient,
		fisuEndpoint:     fisuEndpoint,
		worldsPopulation: containers.NewExpiableValue[WorldsPopulation](time.Minute),
	}
}

func (c *Client) Start() {
	go c.worldsPopulation.StartExpiration()
}

func (c *Client) Stop() {
	c.worldsPopulation.StopExpiration()
}

func (c *Client) Endpoint() string {
	return c.fisuEndpoint
}

func (c *Client) WorldsPopulation(ctx context.Context) (WorldsPopulation, error) {
	return c.worldsPopulation.Load(func() (WorldsPopulation, error) {
		url := c.fisuEndpoint + populationApiUrl
		var contentBody Response[WorldsPopulation]
		err := httpx.GetJson(ctx, c.httpClient, url, &contentBody)
		if err != nil {
			return WorldsPopulation{}, err
		}
		return contentBody.Result, nil
	})
}
