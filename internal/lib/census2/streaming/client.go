package streaming

import (
	"context"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/mitchellh/mapstructure"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/messages"
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

type EventType string

type Event pubsub.Event[EventType]

const (
	MessageReceivedEventType EventType = "messageReceived"
)

type MessageReceived map[string]any

func (MessageReceived) Type() EventType {
	return MessageReceivedEventType
}

type Client struct {
	endpoint                 string
	env                      string
	serviceId                string
	conn                     *websocket.Conn
	msgBuffer                core.MessageBase
	connStateChangeMsgBuffer messages.ConnectionStateChanged
	connectionTimeout        time.Duration
	publisher                pubsub.Publisher[EventType]
}

func NewClient(endpoint string, env string, serviceId string, publisher pubsub.Publisher[EventType]) *Client {
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

func (c *Client) checkConnectionStateChanged(msg MessageReceived) error {
	err := core.AsMessageBase(msg, &c.msgBuffer)
	if err != nil {
		return err
	}
	if !messages.IsConnectionStateChangedMessage(c.msgBuffer) {
		return ErrInvalidConnectionMessage
	}
	err = mapstructure.Decode(msg, &c.connStateChangeMsgBuffer)
	if err != nil {
		return err
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

	var data map[string]any
	if err = wsjson.Read(ctxWithTimeout, conn, &data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = c.checkConnectionStateChanged(data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.conn = conn
	// Skip the write of connection message cause this write
	// can lock the execution in unexpected place.
	//
	// Non blocking write (drop msg if buffer is full)
	// is the same as no write at all.
	//
	// c.Msg <- data
	return nil
}

func (c *Client) Subscribe(ctx context.Context, settings commands.SubscriptionSettings) error {
	const op = "census2.streaming.Client.Subscribe"
	err := wsjson.Write(ctx, c.conn, commands.Subscribe(settings))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for {
		var data interface{}
		if err := wsjson.Read(ctx, c.conn, &data); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		msg, ok := data.(MessageReceived)
		if !ok {
			// TODO: Use optional unknown message publisher
			continue
		}
		if err := c.checkConnectionStateChanged(msg); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err = c.publisher.Publish(msg); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
}

func (c *Client) Close() error {
	defer func() {
		c.conn = nil
	}()
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
