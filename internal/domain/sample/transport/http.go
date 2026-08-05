package transport

import (
	"clean-template/internal/domain/sample/handler"

	"github.com/go-chi/chi/v5"
)

func RegisterHTTP(r chi.Router, h *handler.HTTPHandler) {
	r.Route("/samples", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}
