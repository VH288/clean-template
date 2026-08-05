package helper

import (
	"context"

	"clean-template/internal/pkg/constant"
)

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(constant.ContextKeyRequestID).(string); ok {
		return v
	}
	return ""
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, constant.ContextKeyRequestID, requestID)
}
