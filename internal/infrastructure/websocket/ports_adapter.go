package websocket

import (
	"context"
	"net/http"

	"clean-template/internal/domain/ports"

	"github.com/coder/websocket"
)

type connAdapter struct {
	conn *websocket.Conn
}

func (c connAdapter) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.Read(ctx)
	return data, err
}

// PortsAdapter exposes Hub through domain ports.WebSocketHub.
type PortsAdapter struct {
	hub *Hub
}

func NewPortsAdapter(hub *Hub) *PortsAdapter {
	return &PortsAdapter{hub: hub}
}

func (a *PortsAdapter) Upgrade(w http.ResponseWriter, r *http.Request) (ports.WebSocketConn, error) {
	conn, err := a.hub.Upgrade(w, r)
	if err != nil {
		return nil, err
	}
	return connAdapter{conn: conn}, nil
}

func (a *PortsAdapter) Unregister(conn ports.WebSocketConn) {
	if c, ok := conn.(connAdapter); ok {
		a.hub.Unregister(c.conn)
	}
}

func (a *PortsAdapter) WriteJSON(ctx context.Context, conn ports.WebSocketConn, payload any) error {
	if c, ok := conn.(connAdapter); ok {
		return WriteJSON(ctx, c.conn, payload)
	}
	return nil
}

var _ ports.WebSocketHub = (*PortsAdapter)(nil)
