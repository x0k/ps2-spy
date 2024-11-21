package census2

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"
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

func (c *Client) ExecutePrepared(ctx context.Context, collection, url string) ([]any, error) {
	const op = "census2.Client.ExecutePrepared"
	content, err := httpx.GetJson[map[string]any](ctx, c.httpClient, url)
	if err != nil {
		return nil, err
	}
	data, ok := content[collection+"_list"].([]any)
	if !ok {
		return nil, fmt.Errorf("%s decoding %v: %w", op, content, ErrFailedToDecode)
	}
	return data, nil
}

func (c *Client) Execute(ctx context.Context, q *Query) ([]any, error) {
	return c.ExecutePrepared(ctx, q.Collection(), c.ToURL(q))
}

type DecodeError[T any] struct {
	Index int
	Item  T
	Err   error
}

func (e *DecodeError[T]) Error() string {
	return fmt.Sprintf("item[%d] %v: %s", e.Index, e.Item, e.Err.Error())
}

type DecodeErrors[T any] []DecodeError[T]

func (e DecodeErrors[T]) Error() string {
	var builder strings.Builder
	builder.WriteString("failed to decode:\n")
	builder.WriteString(e[0].Error())
	for i := 1; i < len(e); i++ {
		builder.WriteByte('\n')
		builder.WriteString(e[i].Error())
	}
	return builder.String()
}

func DecodeCollection[T any](items []any) ([]T, error) {
	res := make([]T, len(items))
	errs := make([]DecodeError[T], 0, len(items))
	for i, item := range items {
		err := mapstructure.Decode(item, &res[i])
		if err != nil {
			errs = append(errs, DecodeError[T]{Index: i, Item: res[i], Err: err})
		}
	}
	if len(errs) > 0 {
		return res, DecodeErrors[T](errs)
	}
	return res, nil
}

// Provided type `T` should have `mapstructure` tags
func ExecuteAndDecode[T any](ctx context.Context, c *Client, q *Query) ([]T, error) {
	data, err := c.Execute(ctx, q)
	if err != nil {
		return nil, err
	}
	return DecodeCollection[T](data)
}

func ExecutePreparedAndDecode[T any](ctx context.Context, c *Client, collection, url string) ([]T, error) {
	data, err := c.ExecutePrepared(ctx, collection, url)
	if err != nil {
		return nil, err
	}
	return DecodeCollection[T](data)
}
