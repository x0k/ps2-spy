package population

import (
	"context"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
	"github.com/x0k/ps2-spy/internal/httpx"
)

type Client struct {
	httpClient       *http.Client
	endpoint         string
	worldsPopulation *containers.ExpiableValue[[]WorldPopulation]
}

const populationAllUrl = "/population/all"

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		worldsPopulation: containers.NewExpiableValue[[]WorldPopulation](
			time.Minute,
		),
	}
}

func (c *Client) Start() {
	go c.worldsPopulation.StartExpiration()
}

func (c *Client) Stop() {
	c.worldsPopulation.StopExpiration()
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) AllPopulation(ctx context.Context) ([]WorldPopulation, error) {
	return c.worldsPopulation.Load(func() ([]WorldPopulation, error) {
		return httpx.GetJson[[]WorldPopulation](ctx, c.httpClient, c.endpoint+populationAllUrl)
	})
}
