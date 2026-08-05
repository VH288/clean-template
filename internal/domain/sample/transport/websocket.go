package transport

import (
	"clean-template/internal/domain/sample/handler"

	"github.com/go-chi/chi/v5"
)

func RegisterWebSocket(r chi.Router, h *handler.WSHandler) {
	r.Get("/ws/samples", h.ServeHTTP)
}
