package census2

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

type Client struct {
	httpClient     *http.Client
	censusEndpoint string
	serviceId      string
	cache          *expirable.LRU[string, []any]
}

func NewClient(censusEndpoint string, serviceId string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:     httpClient,
		censusEndpoint: censusEndpoint,
		serviceId:      serviceId,
		cache:          expirable.NewLRU[string, []any](100, nil, time.Minute),
	}
}

func (c *Client) Endpoint() string {
	return c.censusEndpoint
}

func (c *Client) Execute(ctx context.Context, q *Query) ([]any, error) {
	builder := strings.Builder{}
	builder.WriteString(c.censusEndpoint)
	if c.serviceId != "" {
		builder.WriteString("/s:")
		builder.WriteString(c.serviceId)
	}
	builder.WriteString("/json/")
	q.print(&builder)
	url := builder.String()
	if cached, ok := c.cache.Get(url); ok {
		return cached, nil
	}
	content, err := httpx.GetJson[map[string]any](ctx, c.httpClient, url)
	if err != nil {
		return nil, err
	}
	propertyIndex := fmt.Sprintf("%s_list", q.Collection())
	data := content[propertyIndex].([]any)
	c.cache.Add(url, data)
	return data, nil
}

func ExecuteAndDecode[T any](ctx context.Context, c *Client, q *Query) ([]T, error) {
	data, err := c.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	items := make([]T, len(data))
	for i, item := range data {
		mapstructure.Decode(item, &items[i])
	}
	return items, nil
}
