package census2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/httpx"
)

var ErrFailedToDecode = fmt.Errorf("failed to decode")

type Client struct {
	httpClient     *http.Client
	censusEndpoint string
	serviceId      string
}

func NewClient(censusEndpoint string, serviceId string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:     httpClient,
		censusEndpoint: censusEndpoint,
		serviceId:      serviceId,
	}
}

func (c *Client) Endpoint() string {
	return c.censusEndpoint
}

func (c *Client) ToURL(q *Query) string {
	builder := strings.Builder{}
	builder.WriteString(c.censusEndpoint)
	if c.serviceId != "" {
		builder.WriteString("/s:")
		builder.WriteString(c.serviceId)
	}
	// builder.WriteString("/json/")
	builder.WriteByte('/')
	q.print(&builder)
	return builder.String()
}

func (c *Client) ExecutePrepared(ctx context.Context, collection, url string) (json.RawMessage, error) {
	const op = "census2.Client.ExecutePrepared"
	content, err := httpx.GetJson[map[string]json.RawMessage](ctx, c.httpClient, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	data, ok := content[collection+"_list"]
	if !ok {
		return nil, fmt.Errorf("%s decoding %v: %w", op, content, ErrFailedToDecode)
	}
	return data, nil
}

func (c *Client) Execute(ctx context.Context, q *Query) (json.RawMessage, error) {
	return c.ExecutePrepared(ctx, q.Collection(), c.ToURL(q))
}

func DecodeCollection[T any](items json.RawMessage) ([]T, error) {
	var res []T
	err := json.Unmarshal(items, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}
	return res, nil
}

func ExecuteAndDecode[T any](ctx context.Context, c *Client, q *Query) ([]T, error) {
	data, err := c.Execute(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to execute: %w", err)
	}
	return DecodeCollection[T](data)
}

func ExecutePreparedAndDecode[T any](ctx context.Context, c *Client, collection, url string) ([]T, error) {
	data, err := c.ExecutePrepared(ctx, collection, url)
	if err != nil {
		return nil, fmt.Errorf("failed to execute prepared: %w", err)
	}
	return DecodeCollection[T](data)
}
