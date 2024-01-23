package ps2alerts

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

const alertsUrl = "/instances/active"

type Client struct {
	httpClient *http.Client
	endpoint   string
	alerts     *containers.Expiable[[]Alert]
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		alerts:     containers.NewExpiable[[]Alert](time.Minute),
	}
}

func (c *Client) Start(ctx context.Context, wg *sync.WaitGroup) {
	c.alerts.Start(ctx, wg)
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) Alerts(ctx context.Context) ([]Alert, error) {
	return c.alerts.Load(func() ([]Alert, error) {
		url := c.endpoint + alertsUrl
		return httpx.GetJson[[]Alert](ctx, c.httpClient, url)
	})
}
