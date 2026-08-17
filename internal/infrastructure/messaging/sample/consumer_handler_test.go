package sample_test

import (
	"fmt"
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

type memIdempotency struct {
	seen map[string]struct{}
}

func (m *memIdempotency) IsProcessed(_ context.Context, eventID string) (bool, error) {
	_, ok := m.seen[eventID]
	return ok, nil
}

func (m *memIdempotency) MarkProcessed(_ context.Context, eventID string) error {
	if m.seen == nil {
		m.seen = map[string]struct{}{}
	}
	m.seen[eventID] = struct{}{}
	return nil
}

type noopDLQ struct{}

func (noopDLQ) PublishDLQ(context.Context, string, []byte, string) error { return nil }

func TestEventHandler_Created(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger(), &memIdempotency{}, noopDLQ{})
	s := &entity.Sample{ID: "id-1", Name: "alpha", Status: entity.StatusActive}
	payload, err := json.Marshal(event.Envelope{EventID: "evt-1", Type: event.TypeCreated, Sample: s})
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
}

func TestEventHandler_InvalidJSON(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger(), &memIdempotency{}, noopDLQ{})
	err := h.Handle(context.Background(), nil, []byte("not-json"))
	require.Error(t, err)
	require.True(t, samplemsg.IsPoisonError(err))
}

func TestEventHandler_UnknownType(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger(), &memIdempotency{}, noopDLQ{})
	s := &entity.Sample{ID: "id-1", Name: "alpha"}
	payload, err := json.Marshal(event.Envelope{EventID: "evt-2", Type: "sample.unknown", Sample: s})
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
}

func TestEventHandler_Idempotent(t *testing.T) {
	store := &memIdempotency{}
	h := samplemsg.NewEventHandler(testLogger(), store, noopDLQ{})
	s := &entity.Sample{ID: "id-1", Name: "alpha", Status: entity.StatusActive}
	payload, err := json.Marshal(event.Envelope{EventID: "evt-3", Type: event.TypeCreated, Sample: s})
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
	require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
}

func TestEventHandler_UpdatedAndDeleted(t *testing.T) {
	h := samplemsg.NewEventHandler(testLogger(), &memIdempotency{}, noopDLQ{})
	now := time.Now().UTC()
	s := &entity.Sample{ID: "id-1", Name: "alpha", Status: entity.StatusInactive, UpdatedAt: now}

	for i, typ := range []string{event.TypeUpdated, event.TypeDeleted} {
		payload, err := json.Marshal(event.Envelope{EventID: fmt.Sprintf("evt-%d", i), Type: typ, Sample: s})
		require.NoError(t, err)
		require.NoError(t, h.Handle(context.Background(), []byte("id-1"), payload))
	}
}
