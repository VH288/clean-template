package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"clean-template/internal/domain/healthcheck"
	"clean-template/internal/domain/ports"
	"clean-template/internal/pkg/constant"
	"clean-template/internal/pkg/formatter"
	healthv1 "clean-template/internal/proto/healthcheck/v1"
)

type HTTPHandler struct {
	usecase healthcheck.Usecase
}

func NewHTTPHandler(usecase healthcheck.Usecase) *HTTPHandler {
	return &HTTPHandler{usecase: usecase}
}

func (h *HTTPHandler) Check(w http.ResponseWriter, r *http.Request) {
	report := h.usecase.CheckHTTP(r.Context())
	status := http.StatusOK
	if report.Status != constant.HealthStatusUP {
		status = http.StatusServiceUnavailable
	}
	formatter.Success(w, status, "healthcheck", report)
}

type GRPCHandler struct {
	healthv1.UnimplementedHealthcheckServiceServer
	usecase healthcheck.Usecase
}

func NewGRPCHandler(usecase healthcheck.Usecase) *GRPCHandler {
	return &GRPCHandler{usecase: usecase}
}

func (h *GRPCHandler) Check(ctx context.Context, _ *healthv1.CheckRequest) (*healthv1.CheckResponse, error) {
	report := h.usecase.CheckGRPC(ctx)
	resp := &healthv1.CheckResponse{Status: report.Status}
	for _, d := range report.Dependencies {
		resp.Dependencies = append(resp.Dependencies, &healthv1.Dependency{
			Name:    d.Name,
			Status:  d.Status,
			Message: d.Message,
		})
	}
	return resp, nil
}

type WSHandler struct {
	usecase healthcheck.Usecase
	hub     ports.WebSocketHub
}

func NewWSHandler(usecase healthcheck.Usecase, hub ports.WebSocketHub) *WSHandler {
	return &WSHandler{usecase: usecase, hub: hub}
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.hub.Upgrade(w, r)
	if err != nil {
		return
	}
	defer h.hub.Unregister(conn)

	ctx := r.Context()
	for {
		data, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var msg struct {
			Action string `json:"action"`
		}
		_ = json.Unmarshal(data, &msg)
		if msg.Action == "" {
			msg.Action = "ping"
		}

		report := h.usecase.CheckWebSocket(ctx)
		_ = h.hub.WriteJSON(ctx, conn, map[string]any{
			"action": msg.Action,
			"data":   report,
		})
	}
}
