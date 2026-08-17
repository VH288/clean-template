package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"clean-template/internal/domain/ports"
	"clean-template/internal/domain/sample"
	"clean-template/internal/domain/sample/dto"
	"clean-template/internal/domain/sample/mapper"
	"clean-template/internal/pkg/validator"
)

type WSHandler struct {
	usecase sample.Usecase
	hub     ports.WebSocketHub
}

func NewWSHandler(usecase sample.Usecase, hub ports.WebSocketHub) *WSHandler {
	return &WSHandler{usecase: usecase, hub: hub}
}

type wsEnvelope struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
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

		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			_ = h.hub.WriteJSON(ctx, conn, map[string]any{"error": "invalid payload"})
			continue
		}

		resp, err := h.dispatch(ctx, env)
		if err != nil {
			_ = h.hub.WriteJSON(ctx, conn, map[string]any{"error": err.Error()})
			continue
		}
		_ = h.hub.WriteJSON(ctx, conn, resp)
	}
}

func (h *WSHandler) dispatch(ctx context.Context, env wsEnvelope) (any, error) {
	switch env.Action {
	case "create":
		var req dto.CreateSampleRequest
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return nil, err
		}
		if err := validator.Struct(req); err != nil {
			return nil, err
		}
		s, err := h.usecase.Create(ctx, req.Name, req.Description, req.Status)
		if err != nil {
			return nil, err
		}
		return map[string]any{"action": "create", "data": mapper.ToSampleResponse(s)}, nil
	case "get":
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return nil, err
		}
		s, err := h.usecase.GetByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"action": "get", "data": mapper.ToSampleResponse(s)}, nil
	case "list":
		var req dto.ListSampleQuery
		_ = json.Unmarshal(env.Payload, &req)
		items, total, err := h.usecase.List(ctx, req.Page, req.PerPage)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"action": "list",
			"data":   mapper.ToSampleResponses(items),
			"total":  total,
		}, nil
	case "update":
		var req struct {
			ID string `json:"id"`
			dto.UpdateSampleRequest
		}
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return nil, err
		}
		if err := validator.Struct(req.UpdateSampleRequest); err != nil {
			return nil, err
		}
		s, err := h.usecase.Update(ctx, req.ID, req.Name, req.Description, req.Status)
		if err != nil {
			return nil, err
		}
		return map[string]any{"action": "update", "data": mapper.ToSampleResponse(s)}, nil
	case "delete":
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return nil, err
		}
		if err := h.usecase.Delete(ctx, req.ID); err != nil {
			return nil, err
		}
		return map[string]any{"action": "delete", "success": true}, nil
	default:
		return nil, fmt.Errorf("unknown action: %s", env.Action)
	}
}
