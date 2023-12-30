package httpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func GetJson(ctx context.Context, client *http.Client, url string, v any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
