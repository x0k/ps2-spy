package streaming

import "context"

type JSONConn interface {
	ReadJSON(ctx context.Context, v any) error
	WriteJSON(ctx context.Context, v any) error
	Close(code int, reason string) error
}

type Dialer interface {
	Dial(ctx context.Context, urlStr string) (JSONConn, error)
}
