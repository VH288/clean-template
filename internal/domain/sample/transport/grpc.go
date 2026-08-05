package transport

import (
	"clean-template/internal/domain/sample/handler"
	samplev1 "clean-template/internal/proto/sample/v1"

	"google.golang.org/grpc"
)

func RegisterGRPC(server *grpc.Server, h *handler.GRPCHandler) {
	samplev1.RegisterSampleServiceServer(server, h)
}
