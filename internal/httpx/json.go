package httpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func GetJson[T any](ctx context.Context, client *http.Client, url string) (T, error) {
	req, err := http.NewRequest("GET", url, nil)
	var v T
	if err != nil {
		return v, err
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(body, &v)
	return v, err
}
