package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrUnexpectedStatusCode = errors.New("unexpected HTTP status code")

func GetJson[T any](ctx context.Context, client *http.Client, url string) (T, error) {
	req, err := http.NewRequest("GET", url, nil)
	var v T
	if err != nil {
		return v, fmt.Errorf("failed to create request: %w", err)
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return v, fmt.Errorf("%w: %s", ErrUnexpectedStatusCode, resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&v); err != nil {
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	return v, nil
}
