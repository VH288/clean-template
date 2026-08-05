package transport

import (
	"clean-template/internal/domain/healthcheck/handler"
	healthv1 "clean-template/internal/proto/healthcheck/v1"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

func RegisterHTTP(r chi.Router, h *handler.HTTPHandler) {
	r.Get("/health", h.Check)
	r.Get("/healthz", h.Check)
}

func RegisterGRPC(server *grpc.Server, h *handler.GRPCHandler) {
	healthv1.RegisterHealthcheckServiceServer(server, h)
}

func RegisterWebSocket(r chi.Router, h *handler.WSHandler) {
	r.Get("/ws/health", h.ServeHTTP)
}
