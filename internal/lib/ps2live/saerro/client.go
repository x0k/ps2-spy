package saerro

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient          *http.Client
	endpoint            string
	allWorldsPopulation *containers.Expiable[AllWorldsPopulation]
}

const graphqlUrl = "/graphql"

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		allWorldsPopulation: containers.NewExpiable[AllWorldsPopulation](
			time.Minute,
		),
	}
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) Start(ctx context.Context, wg *sync.WaitGroup) {
	c.allWorldsPopulation.Start(ctx, wg)
}

const allWorldsPopulationQuery = graphqlUrl + "?query={allWorlds{id,name,zones{all{id,name,population{total,nc,vs,tr,ns}}}}}"

func (c *Client) AllWorldsPopulation(ctx context.Context) (AllWorldsPopulation, error) {
	return c.allWorldsPopulation.Load(func() (AllWorldsPopulation, error) {
		res, err := httpx.GetJson[GraphqlResponse[AllWorldsPopulation]](ctx, c.httpClient, c.endpoint+allWorldsPopulationQuery)
		if err != nil {
			return AllWorldsPopulation{}, err
		}
		return res.Data, nil
	})
}
