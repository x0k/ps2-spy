package census2

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/x0k/ps2-spy/internal/httpx"
)

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

func (c *Client) Execute(ctx context.Context, q Query) ([]any, error) {
	builder := strings.Builder{}
	builder.WriteString(c.censusEndpoint)
	builder.WriteString("s:")
	builder.WriteString(c.serviceId)
	builder.WriteString("/json/")
	q.print(&builder)
	url := builder.String()
	content, err := httpx.GetJson[map[string]any](ctx, c.httpClient, url)
	if err != nil {
		return nil, err
	}
	propertyIndex := fmt.Sprintf("%s_list", q.Collection())
	return content[propertyIndex].([]any), nil
}
