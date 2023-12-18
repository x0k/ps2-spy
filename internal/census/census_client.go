package census

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBatchLimit = 500

// CensusClient is an object that retains a client and census configuration states between generated queries
type CensusClient struct {
	serviceID           string
	collectionNamespace string
	client              *http.Client
}

// NewCensusClient creates a CensusClient object
func NewCensusClient(serviceID string, collectionNamespace string) *CensusClient {
	return &CensusClient{
		serviceID:           serviceID,
		collectionNamespace: collectionNamespace,
		client:              &http.Client{},
	}
}

func (c *CensusClient) executeQuery(query *Query) ([]any, error) {
	requestURL := c.createRequestURL(query)

	resp, err := c.client.Get(requestURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var contentBody map[string]any
	err = json.Unmarshal(body, &contentBody)
	if err != nil {
		return nil, err
	}

	propertyIndex := fmt.Sprintf("%s_list", query.Collection)
	return contentBody[propertyIndex].([]any), nil
}

func (c *CensusClient) executeQueryBatch(query *Query) ([]any, error) {
	count := 0

	batchResult := make([]any, 0)

	if query.Limit <= 0 {
		query.SetLimit(defaultBatchLimit)
	}

	if query.Start < 0 {
		query.SetStart(count)
	}

	result, err := c.executeQuery(query)
	if err != nil {
		return nil, err
	}

	if len(result) < query.Limit {
		return result, nil
	}

	for ok := true; ok; ok = len(result) > 0 {
		batchResult = append(batchResult, result...)

		if len(result) < query.Limit {
			return batchResult, nil
		}

		count += len(result)
		query.SetStart(count)

		result, err = c.executeQuery(query)
		if err != nil {
			return nil, err
		}
	}

	return batchResult, nil
}

func (c *CensusClient) createRequestURL(query *Query) string {
	sID := c.serviceID
	ns := c.collectionNamespace

	encArgs := query.String()
	return endpointCollectionGet(sID, ns, encArgs)
}
