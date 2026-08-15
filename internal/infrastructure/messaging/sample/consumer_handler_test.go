package sample_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/event"
	samplemsg "clean-template/internal/infrastructure/messaging/sample"

	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestEventHandler_Created(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger())
	s := &entity.Sample{ID: "id-1", Name: "alpha", Status: entity.StatusActive}
	payload, err := json.Marshal(event.Envelope{Type: event.TypeCreated, Sample: s})
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
}

func TestEventHandler_InvalidJSON(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger())
	err := h.Handle(context.Background(), nil, []byte("not-json"))
	require.Error(t, err)
}

func TestEventHandler_UnknownType(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger())
	s := &entity.Sample{ID: "id-1", Name: "alpha"}
	payload, err := json.Marshal(event.Envelope{Type: "sample.unknown", Sample: s})
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
}

func TestEventHandler_UpdatedAndDeleted(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger())
	now := time.Now().UTC()
	s := &entity.Sample{ID: "id-1", Name: "alpha", Status: entity.StatusInactive, UpdatedAt: now}

	for _, typ := range []string{event.TypeUpdated, event.TypeDeleted} {
		payload, err := json.Marshal(event.Envelope{Type: typ, Sample: s})
		require.NoError(t, err)
		require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
	}
}
