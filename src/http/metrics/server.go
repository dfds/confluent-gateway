package metrics

import (
	"context"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type Server struct {
	logger logging.Logger
	server *http.Server
}

func NewServer(logger logging.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	return &Server{
		logger: logger,
		server: server,
	}
}

func (s *Server) Open() error {
	s.logger.Debug("Starting HTTP Server")

	if err := s.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}
	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s.logger.Debug("Shutting down HTTP Server")

	return s.server.Shutdown(ctx)
}
