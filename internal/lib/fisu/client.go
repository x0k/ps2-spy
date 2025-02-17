package fisu

import (
	"context"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient   *http.Client
	fisuEndpoint string
}

const populationApiUrl = "/api/population/?world=1,10,13,17,19,24,40,1000,2000"

func NewClient(fisuEndpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:   httpClient,
		fisuEndpoint: fisuEndpoint,
	}
}

func (c *Client) Endpoint() string {
	return c.fisuEndpoint
}

func (c *Client) WorldsPopulation(ctx context.Context) (WorldsPopulation, error) {
	url := c.fisuEndpoint + populationApiUrl
	res, err := httpx.GetJson[Response[WorldsPopulation]](ctx, c.httpClient, url)
	if err != nil {
		return WorldsPopulation{}, err
	}
	return res.Result, nil
}
