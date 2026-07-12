package grpc

import (
	"context"
	"fmt"
	"net"

	grpchandler "clean-template/internal/delivery/grpc/handler"
	"clean-template/internal/delivery/grpc/interceptor"
	samplev1 "clean-template/proto/sample/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	port          string
	log           *logrus.Logger
	sampleHandler *grpchandler.SampleHandler
	server        *grpc.Server
}

func NewServer(port string, log *logrus.Logger, sampleHandler *grpchandler.SampleHandler) *Server {
	return &Server{
		port:          port,
		log:           log,
		sampleHandler: sampleHandler,
	}
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("listen grpc port: %w", err)
	}

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.RequestLogger(s.log)),
	)
	samplev1.RegisterSampleServiceServer(s.server, s.sampleHandler)

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("start listening grpc on port: ", s.port)
		if err := s.server.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down grpc server")
		s.server.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}
