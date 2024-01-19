package streaming

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mitchellh/mapstructure"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/core"
	ps2messages "github.com/x0k/ps2-spy/internal/lib/census2/streaming/messages"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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
	log                      *slog.Logger
	endpoint                 string
	env                      string
	serviceId                string
	conn                     *websocket.Conn
	msgBuffer                core.MessageBase
	connStateChangeMsgBuffer ps2messages.ConnectionStateChanged
	connectionTimeout        time.Duration
	Msg                      chan map[string]any
}

func NewClient(log *slog.Logger, endpoint string, env string, serviceId string) *Client {
	return &Client{
		log: log.With(
			slog.String("component", "census2.streaming.Client"),
			slog.String("endpoint", endpoint),
			slog.String("env", env),
		),
		endpoint:          endpoint,
		env:               env,
		serviceId:         serviceId,
		connectionTimeout: time.Duration(10) * time.Second,
		Msg:               make(chan map[string]any),
	}
}

func (c *Client) Environment() string {
	return c.env
}

func (c *Client) fillConnectionStateChangedBuffer(msg map[string]any) error {
	err := core.AsMessageBase(msg, &c.msgBuffer)
	if err != nil {
		return err
	}
	if !ps2messages.IsConnectionStateChangedMessage(c.msgBuffer) {
		return ErrInvalidConnectionMessage
	}
	err = mapstructure.Decode(msg, &c.connStateChangeMsgBuffer)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Connect(ctx context.Context) error {
	const op = "census2.streaming.Client.Connect"
	c.log.Info("connecting to websocket")
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
	timeout := time.AfterFunc(c.connectionTimeout, func() {
		conn.Close(websocket.StatusNormalClosure, "connection timeout")
	})
	defer timeout.Stop()
	var data map[string]any
	err = wsjson.Read(ctx, conn, &data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	// TODO: check that timeout was not triggered before assigning connection
	timeout.Stop()
	err = c.fillConnectionStateChangedBuffer(data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if c.connStateChangeMsgBuffer.Connected != core.True {
		return fmt.Errorf("%s: %w", op, ErrConnectionFailed)
	}
	c.conn = conn
	// Skip the write of connection message cause this write
	// can lock the execution in unexpected place.
	//
	// Non blocking write (drop msg if buffer is full)
	// is the same as no write at all.
	//
	// c.Msg <- data
	c.log.Info("connected")
	return nil
}

func (c *Client) onMessage(msg any) error {
	m, ok := msg.(map[string]any)
	if !ok {
		c.log.Warn("unexpected message type", slog.Any("msg", msg))
		return nil
	}
	if c.fillConnectionStateChangedBuffer(m) == nil && c.connStateChangeMsgBuffer.Connected == core.False {
		c.log.Info("disconnected by server")
		return ErrDisconnectedByServer
	}
	// Lock for backpressure
	c.Msg <- m
	return nil
}

func (c *Client) Subscribe(ctx context.Context, settings ps2commands.SubscriptionSettings) error {
	const op = "census2.streaming.Client.Subscribe"
	err := wsjson.Write(ctx, c.conn, ps2commands.Subscribe(settings))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for {
		var data interface{}
		err := wsjson.Read(ctx, c.conn, &data)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		err = c.onMessage(data)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
}

func (c *Client) Close() error {
	c.log.Info("closing websocket connection")
	defer func() {
		c.conn = nil
	}()
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
