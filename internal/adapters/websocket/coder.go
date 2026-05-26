package websocket_adapters

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
)

type coderDialer struct{}

func (coderDialer) Dial(ctx context.Context, urlStr string) (streaming.JSONConn, error) {
	conn, _, err := websocket.Dial(ctx, urlStr, nil)
	if err != nil {
		return nil, err
	}
	return &coderConn{conn: conn}, nil
}

func NewCoderDialer() streaming.Dialer {
	return coderDialer{}
}

type coderConn struct {
	conn *websocket.Conn
}

func (c *coderConn) ReadJSON(ctx context.Context, v any) error {
	return wsjson.Read(ctx, c.conn, v)
}

func (c *coderConn) WriteJSON(ctx context.Context, v any) error {
	return wsjson.Write(ctx, c.conn, v)
}

func (c *coderConn) Close(code int, reason string) error {
	return c.conn.Close(websocket.StatusCode(code), reason)
}
