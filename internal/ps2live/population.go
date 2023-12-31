package ps2live

import (
	"context"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
	"github.com/x0k/ps2-spy/internal/httpx"
)

type PopulationClient struct {
	httpClient       *http.Client
	endpoint         string
	worldsPopulation *containers.ExpiableValue[[]WorldPopulation]
}

const populationAllUrl = "/population/all"

func NewPopulationClient(endpoint string, httpClient *http.Client) *PopulationClient {
	return &PopulationClient{
		httpClient: httpClient,
		endpoint:   endpoint,
		worldsPopulation: containers.NewExpiableValue[[]WorldPopulation](
			time.Minute,
		),
	}
}

func (c *PopulationClient) Start() {
	go c.worldsPopulation.StartExpiration()
}

func (c *PopulationClient) Stop() {
	c.worldsPopulation.StopExpiration()
}

func (c *PopulationClient) Endpoint() string {
	return c.endpoint
}

func (c *PopulationClient) AllPopulation(ctx context.Context) ([]WorldPopulation, error) {
	return c.worldsPopulation.Load(func() ([]WorldPopulation, error) {
		return httpx.GetJson[[]WorldPopulation](ctx, c.httpClient, c.endpoint+populationAllUrl)
	})
}
