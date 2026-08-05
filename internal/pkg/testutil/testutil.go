package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"clean-template/internal/pkg/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewContext() context.Context {
	return helper.WithRequestID(context.Background(), uuid.NewString())
}

func NewContextWithTimeout(t *testing.T, d time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(NewContext(), d)
}

func DecodeJSON(t *testing.T, rr *httptest.ResponseRecorder, dest any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(rr.Body).Decode(dest))
}

func PerformRequest(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func AssertStatus(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()
	require.Equal(t, want, rr.Code)
}
