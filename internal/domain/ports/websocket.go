package ports

import (
	"context"
	"net/http"
)

// WebSocketConn is a minimal connection surface for domain handlers.
type WebSocketConn interface {
	Read(ctx context.Context) ([]byte, error)
}

// WebSocketHub upgrades HTTP connections and writes JSON payloads.
type WebSocketHub interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (WebSocketConn, error)
	Unregister(conn WebSocketConn)
	WriteJSON(ctx context.Context, conn WebSocketConn, payload any) error
}
