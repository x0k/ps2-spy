package census

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type httpCensusClient struct {
	client         *http.Client
	censusEndpoint string
	serviceId      string
}

func NewClient(censusEndpoint string, serviceId string, client *http.Client) CensusClient {
	return &httpCensusClient{
		client:         client,
		censusEndpoint: censusEndpoint,
		serviceId:      serviceId,
	}
}

func (c *httpCensusClient) Execute(query CensusQuery) (any, error) {
	builder := strings.Builder{}
	builder.WriteString(c.censusEndpoint)
	builder.WriteString("s:")
	builder.WriteString(c.serviceId)
	builder.WriteString("/json/")
	query.write(&builder)
	url := builder.String()

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var contentBody map[string]interface{}
	err = json.Unmarshal(body, &contentBody)
	if err != nil {
		return nil, err
	}
	propertyIndex := fmt.Sprintf("%s_list", query.GetCollection())
	return contentBody[propertyIndex].([]interface{}), nil
}
