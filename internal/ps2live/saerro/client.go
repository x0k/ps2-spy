package saerro

import (
	"context"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
	"github.com/x0k/ps2-spy/internal/httpx"
)

type Client struct {
	httpClient          *http.Client
	endpoint            string
	allWorldsPopulation *containers.ExpiableValue[AllWorldsPopulation]
}

const graphqlUrl = "/graphql"

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		allWorldsPopulation: containers.NewExpiableValue[AllWorldsPopulation](
			time.Minute,
		),
	}
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) Start() {
	go c.allWorldsPopulation.StartExpiration()
}

func (c *Client) Stop() {
	c.allWorldsPopulation.StopExpiration()
}

const allWorldsPopulationQuery = graphqlUrl + "?query={allWorlds{id,name,zones{all{id,name,population{total,nc,vs,tr,ns}}}}}"

func (c *Client) AllWorldsPopulation(ctx context.Context) (AllWorldsPopulation, error) {
	return c.allWorldsPopulation.Load(func() (AllWorldsPopulation, error) {
		return httpx.GetJson[AllWorldsPopulation](ctx, c.httpClient, c.endpoint+allWorldsPopulationQuery)
	})
}
