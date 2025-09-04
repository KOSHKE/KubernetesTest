package metrics

import (
	"context"
	"net/http"
	"time"

	"ecommerce-platform/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer provides HTTP endpoint for Prometheus metrics with graceful shutdown
type MetricsServer struct {
	addr   string
	server *http.Server
	logger logger.Logger
}

// MetricsServerConfig holds configuration for MetricsServer
type MetricsServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Handler      http.Handler
	Registry     prometheus.Registerer
	Gatherer     prometheus.Gatherer
}

// DefaultMetricsServerConfig returns default configuration
func DefaultMetricsServerConfig(addr string) *MetricsServerConfig {
	return &MetricsServerConfig{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      nil,
		Registry:     prometheus.DefaultRegisterer,
		Gatherer:     prometheus.DefaultGatherer,
	}
}

// NewMetricsServer creates new metrics server with default configuration
func NewMetricsServer(addr string, logger logger.Logger) *MetricsServer {
	config := DefaultMetricsServerConfig(addr)
	return NewMetricsServerWithConfig(config, logger)
}

// NewMetricsServerWithConfig creates new metrics server with custom configuration
func NewMetricsServerWithConfig(config *MetricsServerConfig, logger logger.Logger) *MetricsServer {
	var handler http.Handler
	if config.Handler != nil {
		// Use custom handler if provided
		handler = config.Handler
	} else {
		// Use default Prometheus handler with OpenMetrics support and metrics instrumentation
		mux := http.NewServeMux()
		metricsHandler := promhttp.InstrumentMetricHandler(
			config.Registry,
			promhttp.HandlerFor(
				config.Gatherer,
				promhttp.HandlerOpts{
					EnableOpenMetrics: true,
				},
			),
		)
		mux.Handle("/metrics", metricsHandler)
		handler = mux
	}

	server := &http.Server{
		Addr:         config.Addr,
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &MetricsServer{
		addr:   config.Addr,
		server: server,
		logger: logger,
	}
}

// Start starts metrics server with custom context for testing and external control
func (s *MetricsServer) Start(ctx context.Context) error {
	// Start server in separate goroutine
	go func() {
		s.logger.Info("metrics server started", "addr", s.addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("failed to start metrics server", "error", err)
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

// Close implements io.Closer interface
func (s *MetricsServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// GetMux returns the HTTP mux for custom handlers (only available when using default handler)
func (s *MetricsServer) GetMux() *http.ServeMux {
	if mux, ok := s.server.Handler.(*http.ServeMux); ok {
		return mux
	}
	return nil
}

// GetServer returns the underlying http.Server for advanced configuration
func (s *MetricsServer) GetServer() *http.Server {
	return s.server
}

// Note: Standard Go and process metrics are registered elsewhere in the project
// to avoid duplicate registration conflicts
