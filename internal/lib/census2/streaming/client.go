package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
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
	log                      *logger.Logger
	endpoint                 string
	env                      string
	serviceId                string
	dialer                   Dialer
	conn                     JSONConn
	connStateChangeMsgBuffer ConnectionStateChanged
	connectionTimeout        time.Duration
	publisher                pubsub.Publisher[json.RawMessage]
}

func NewClient(
	log *logger.Logger,
	endpoint string,
	env string,
	serviceId string,
	publisher pubsub.Publisher[json.RawMessage],
	dialer Dialer,
) *Client {
	return &Client{
		log:               log,
		endpoint:          endpoint,
		env:               env,
		serviceId:         serviceId,
		connectionTimeout: time.Duration(10) * time.Second,
		publisher:         publisher,
		dialer:            dialer,
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
	conn, err := c.dialer.Dial(ctx, c.endpoint+fmt.Sprintf("?environment=%s&service-id=s:%s", c.env, c.serviceId))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	committed := false
	defer func() {
		if !committed {
			if err := conn.Close(1000, "connection failed"); err != nil {
				c.log.Error(ctx, "failed to close websocket connection", sl.Err(err))
			}
		}
	}()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()

	var data json.RawMessage
	if err = conn.ReadJSON(ctxWithTimeout, &data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = c.checkConnectionStateChanged(data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if c.conn != nil {
		if err := c.conn.Close(1000, ""); err != nil {
			c.log.Error(ctx, "failed to close old websocket connection", sl.Err(err))
		}
	}
	c.conn = conn
	committed = true
	return nil
}

func (c *Client) Subscribe(ctx context.Context, settings commands.SubscriptionSettings) error {
	const op = "census2.streaming.Client.Subscribe"
	err := c.conn.WriteJSON(ctx, commands.Subscribe(settings))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for {
		var data json.RawMessage
		if err := c.conn.ReadJSON(ctx, &data); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err := c.checkConnectionStateChanged(data); err == ErrDisconnectedByServer {
			return fmt.Errorf("%s: %w", op, err)
		} else if err == nil {
			continue
		}
		c.publisher.Publish(data)
	}
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	defer func() {
		c.conn = nil
	}()
	return c.conn.Close(1000, "")
}
