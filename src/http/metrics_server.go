package http

import (
	"context"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type MetricsServer struct {
	logger logging.Logger
	server *http.Server
}

func NewMetricsServer(logger logging.Logger) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	return &MetricsServer{
		logger: logger,
		server: server,
	}
}

func (s *MetricsServer) Open() error {
	s.logger.Debug("Starting HTTP MetricsServer")

	if err := s.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}
	return nil
}

func (s *MetricsServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s.logger.Debug("Shutting down HTTP MetricsServer")

	return s.server.Shutdown(ctx)
}
