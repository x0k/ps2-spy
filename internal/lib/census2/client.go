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
	// builder.WriteString("/json/")
	builder.WriteByte('/')
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

type DecodeError struct {
	Index int
	Err   error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("failed to decode item %d: %s", e.Index, e.Err.Error())
}

type DecodeErrors []DecodeError

func (e DecodeErrors) Error() string {
	return fmt.Sprintf("failed to decode %d items", len(e))
}

func DecodeCollection[T any](items []any) ([]T, error) {
	res := make([]T, len(items))
	errs := make([]DecodeError, 0, len(items))
	for i, item := range items {
		err := mapstructure.Decode(item, &res[i])
		if err != nil {
			errs = append(errs, DecodeError{Index: i, Err: err})
		}
	}
	if len(errs) > 0 {
		return res, DecodeErrors(errs)
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
