package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

type HTTPServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewHTTP(addr string, handler http.Handler, logger *slog.Logger, readTimeout, writeTimeout time.Duration) *HTTPServer {
	return &HTTPServer{
		logger: logger,
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
	}
}

func (s *HTTPServer) Start() error {
	s.logger.Info("http server starting", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.server.Shutdown(ctx)
}

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	logger   *slog.Logger
	addr     string
}

func NewGRPC(addr string, server *grpc.Server, logger *slog.Logger) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen grpc: %w", err)
	}
	return &GRPCServer{
		server:   server,
		listener: lis,
		logger:   logger,
		addr:     addr,
	}, nil
}

func (s *GRPCServer) Start() error {
	s.logger.Info("grpc server starting", "addr", s.addr)
	return s.server.Serve(s.listener)
}

func (s *GRPCServer) GracefulStop() {
	s.logger.Info("grpc server shutting down")
	s.server.GracefulStop()
}
