package metrics

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsServer provides HTTP endpoint for Prometheus metrics with graceful shutdown
type MetricsServer struct {
	addr   string
	server *http.Server
	logger *zap.Logger
}

// NewMetricsServer creates new metrics server
func NewMetricsServer(addr string, logger *zap.Logger) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &MetricsServer{
		addr:   addr,
		server: server,
		logger: logger,
	}
}

// NewMetricsServerWithMetrics creates new metrics server with custom PrometheusMetrics
func NewMetricsServerWithMetrics(addr string, logger *zap.Logger, promMetrics *PrometheusMetrics) *MetricsServer {
	mux := http.NewServeMux()

	// Use global registry since PrometheusMetrics uses promauto
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &MetricsServer{
		addr:   addr,
		server: server,
		logger: logger,
	}
}

// Start starts metrics server with graceful shutdown support
func (s *MetricsServer) Start() error {
	// Channel for receiving shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in separate goroutine
	go func() {
		s.logger.Info("metrics server started", zap.String("addr", s.addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("failed to start metrics server", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-stop
	s.logger.Info("received shutdown signal, stopping metrics server")

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	s.logger.Info("metrics server stopped gracefully")
	return nil
}

// StartWithContext starts metrics server with custom context for testing and external control
func (s *MetricsServer) StartWithContext(ctx context.Context) error {
	// Start server in separate goroutine
	go func() {
		s.logger.Info("metrics server started", zap.String("addr", s.addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("failed to start metrics server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	s.logger.Info("context cancelled, stopping metrics server")

	// Graceful shutdown using the same context for proper timeout and cancellation handling
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	s.logger.Info("metrics server stopped gracefully")
	return nil
}

// Shutdown gracefully shuts down the metrics server
func (s *MetricsServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// GetMux returns the HTTP mux for custom handlers
func (s *MetricsServer) GetMux() *http.ServeMux {
	return s.server.Handler.(*http.ServeMux)
}
