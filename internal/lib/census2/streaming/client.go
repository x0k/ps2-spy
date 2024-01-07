package streaming

import (
	"context"
	"fmt"

	"nhooyr.io/websocket"
)

// wss://push.planetside2.com/streaming?environment=[ps2|ps2ps4us|ps2ps4eu]&service-id=s:[your service id]

const (
	Ps2_env      = "ps2"
	Ps2ps4us_env = "ps2ps4us"
	Ps2ps4eu_env = "ps2ps4eu"
)

type Client struct {
	endpoint  string
	env       string
	serviceId string
	conn      *websocket.Conn
}

func NewClient(endpoint string, env string, serviceId string) *Client {
	return &Client{
		endpoint:  endpoint,
		env:       env,
		serviceId: serviceId,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	conn, _, err := websocket.Dial(ctx, c.endpoint+fmt.Sprintf("?environment=%s&service-id=s:%s", c.env, c.serviceId), nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
