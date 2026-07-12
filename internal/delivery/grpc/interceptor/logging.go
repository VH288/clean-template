package interceptor

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func RequestLogger(log *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		log.WithFields(logrus.Fields{
			"method":   info.FullMethod,
			"duration": time.Since(start).String(),
			"error":    err,
		}).Info("grpc request")
		return resp, err
	}
}
