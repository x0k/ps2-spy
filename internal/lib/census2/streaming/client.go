package streaming

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// wss://push.planetside2.com/streaming?environment=[ps2|ps2ps4us|ps2ps4eu]&service-id=s:[your service id]

const (
	Ps2_env      = "ps2"
	Ps2ps4us_env = "ps2ps4us"
	Ps2ps4eu_env = "ps2ps4eu"
)

type eventHandlerInstance struct {
	eventHandler ps2events.EventHandler
}

type Client struct {
	log             *slog.Logger
	endpoint        string
	env             string
	serviceId       string
	conn            *websocket.Conn
	eventHandlersMu sync.RWMutex
	eventHandlers   map[string][]*eventHandlerInstance
}

func NewClient(log *slog.Logger, endpoint string, env string, serviceId string) *Client {
	return &Client{
		log:           log.With(slog.String("component", "census2.streaming.Client")),
		endpoint:      endpoint,
		env:           env,
		serviceId:     serviceId,
		eventHandlers: map[string][]*eventHandlerInstance{},
	}
}

func (c *Client) Connect(ctx context.Context) error {
	const op = "census2.streaming.Client.Connect"
	conn, _, err := websocket.Dial(ctx, c.endpoint+fmt.Sprintf("?environment=%s&service-id=s:%s", c.env, c.serviceId), nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.conn = conn
	return nil
}

func (c *Client) removeEventHandlerInstance(eventType string, instance *eventHandlerInstance) {
	c.eventHandlersMu.Lock()
	defer c.eventHandlersMu.Unlock()
	for i, v := range c.eventHandlers[eventType] {
		if v == instance {
			c.eventHandlers[eventType] = append(c.eventHandlers[eventType][:i], c.eventHandlers[eventType][i+1:]...)
			return
		}
	}
}

func (c *Client) addEventHandler(handler ps2events.EventHandler) func() {
	c.eventHandlersMu.Lock()
	defer c.eventHandlersMu.Unlock()
	instance := &eventHandlerInstance{handler}
	c.eventHandlers[handler.Type()] = append(c.eventHandlers[handler.Type()], instance)
	return func() {
		c.removeEventHandlerInstance(handler.Type(), instance)
	}
}

func (c *Client) AddEventHandler(handler any) func() {
	eventHandler := ps2events.EventHandlerForInterface(handler)

	if eventHandler == nil {
		c.log.Error("invalid event handler type, handler will never be called")
		return func() {}
	}

	return c.addEventHandler(eventHandler)
}

func (c *Client) onMessage(msg any) {
	c.log.Debug("onMessage", slog.Any("msg", msg))
}

func (c *Client) Subscribe(ctx context.Context, settings ps2commands.SubscriptionSettings) error {
	const op = "census2.streaming.Client.Subscribe"
	err := wsjson.Write(ctx, c.conn, settings)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for {
		var data interface{}
		err := wsjson.Read(ctx, c.conn, &data)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		c.onMessage(data)
	}
}

func (c *Client) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
