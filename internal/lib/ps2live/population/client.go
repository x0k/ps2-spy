package population

import (
	"context"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
}

const populationAllUrl = "/population/all"

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) AllPopulation(ctx context.Context) ([]WorldPopulation, error) {
	return httpx.GetJson[[]WorldPopulation](ctx, c.httpClient, c.endpoint+populationAllUrl)
}
