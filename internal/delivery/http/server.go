package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	engine *gin.Engine
	port   string
	log    *logrus.Logger
	server *http.Server
}

func NewServer(engine *gin.Engine, port string, log *logrus.Logger) *Server {
	return &Server{
		engine: engine,
		port:   port,
		log:    log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.port),
		Handler: s.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("start listening http on port: ", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.log.Info("shutting down http server")
		return s.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
