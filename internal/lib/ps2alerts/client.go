package ps2alerts

import (
	"context"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

const alertsUrl = "/instances/active"

type Client struct {
	httpClient *http.Client
	endpoint   string
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) Alerts(ctx context.Context) ([]Alert, error) {
	url := c.endpoint + alertsUrl
	return httpx.GetJson[[]Alert](ctx, c.httpClient, url)
}
