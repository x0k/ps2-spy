package ps2alerts

import (
	"context"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
	"github.com/x0k/ps2-spy/internal/httpx"
)

const alertsUrl = "/instances/active"

type Client struct {
	httpClient *http.Client
	endpoint   string
	alerts     *containers.ExpiableValue[[]Alert]
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
		alerts:     containers.NewExpiableValue[[]Alert](time.Minute),
	}
}

func (c *Client) Start() {
	go c.alerts.StartExpiration()
}

func (c *Client) Stop() {
	c.alerts.StopExpiration()
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
