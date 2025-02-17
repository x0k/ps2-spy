package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

// wss://push.planetside2.com/streaming?environment=[ps2|ps2ps4us|ps2ps4eu]&service-id=s:[your service id]

const (
	Ps2_env      = "ps2"
	Ps2ps4us_env = "ps2ps4us"
	Ps2ps4eu_env = "ps2ps4eu"
)

var ErrUnknownEventHandler = fmt.Errorf("unknown event handler")
var ErrInvalidConnectionMessage = fmt.Errorf("invalid connection message")
var ErrConnectionFailed = fmt.Errorf("failed to connect")
var ErrDisconnectedByServer = fmt.Errorf("disconnected by server")

type Client struct {
	endpoint                 string
	env                      string
	serviceId                string
	conn                     *websocket.Conn
	connStateChangeMsgBuffer ConnectionStateChanged
	connectionTimeout        time.Duration
	publisher                pubsub.Publisher[json.RawMessage]
}

func NewClient(
	endpoint string,
	env string,
	serviceId string,
	publisher pubsub.Publisher[json.RawMessage],
) *Client {
	return &Client{
		endpoint:          endpoint,
		env:               env,
		serviceId:         serviceId,
		connectionTimeout: time.Duration(10) * time.Second,
		publisher:         publisher,
	}
}

func (c *Client) Environment() string {
	return c.env
}

func (c *Client) checkConnectionStateChanged(msg json.RawMessage) error {
	err := json.Unmarshal(msg, &c.connStateChangeMsgBuffer)
	if err != nil {
		return err
	}
	if !IsConnectionStateChangedMessage(c.connStateChangeMsgBuffer.MessageBase) {
		return ErrInvalidConnectionMessage
	}
	if c.connStateChangeMsgBuffer.Connected != core.True {
		return ErrDisconnectedByServer
	}
	return nil
}

func (c *Client) Connect(ctx context.Context) error {
	const op = "census2.streaming.Client.Connect"
	conn, _, err := websocket.Dial(ctx, c.endpoint+fmt.Sprintf("?environment=%s&service-id=s:%s", c.env, c.serviceId), nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if c.conn != nil {
			return
		}
		conn.Close(websocket.StatusNormalClosure, "connection failed")
	}()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()

	var data json.RawMessage
	if err = wsjson.Read(ctxWithTimeout, conn, &data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = c.checkConnectionStateChanged(data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Subscribe(ctx context.Context, settings commands.SubscriptionSettings) error {
	const op = "census2.streaming.Client.Subscribe"
	err := wsjson.Write(ctx, c.conn, commands.Subscribe(settings))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for {
		var data json.RawMessage
		if err := wsjson.Read(ctx, c.conn, &data); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err := c.checkConnectionStateChanged(data); err == ErrDisconnectedByServer {
			return fmt.Errorf("%s: %w", op, err)
		}
		c.publisher.Publish(data)
	}
}

func (c *Client) Close() error {
	defer func() {
		c.conn = nil
	}()
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
