package errors_test

import (
	"net/http"
	"testing"

	apperrors "clean-template/internal/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestAppError_Codes(t *testing.T) {
	require.Equal(t, http.StatusNotFound, apperrors.ErrNotFound.HTTPStatus)
	require.Equal(t, http.StatusBadRequest, apperrors.ErrValidation.HTTPStatus)

	wrapped := apperrors.Wrap(apperrors.ErrNotFound, apperrors.CodeInternal, "boom")
	require.Contains(t, wrapped.Error(), "boom")

	got, ok := apperrors.As(wrapped)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeInternal, got.Code)
}
