package formatter

import (
	"encoding/json"
	"net/http"

	apperrors "clean-template/internal/pkg/errors"
)

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    any         `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, status int, message string, data any) {
	JSON(w, status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithMeta(w http.ResponseWriter, status int, message string, data any, meta *Meta) {
	JSON(w, status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Fail(w http.ResponseWriter, err error) {
	if appErr, ok := apperrors.As(err); ok {
		JSON(w, appErr.HTTPStatus, Response{
			Success: false,
			Error: &ErrorBody{
				Code:    string(appErr.Code),
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}

	JSON(w, http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    string(apperrors.CodeInternal),
			Message: "internal server error",
		},
	})
}
