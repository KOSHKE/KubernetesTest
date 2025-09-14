package metrics

import (
	"time"

	"ecommerce-platform/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// UserMetrics interface defines user service specific metrics
type UserMetrics interface {
	// gRPC metrics
	GRPCRequestDuration(method string, duration time.Duration)
	GRPCRequestTotal(method, status string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// UserPrometheusMetrics implements UserMetrics interface
type UserPrometheusMetrics struct {
	*metrics.PrometheusMetrics

	// gRPC metrics
	grpcRequestTotal    *prometheus.CounterVec
	grpcRequestDuration *prometheus.HistogramVec
}

// NewUserMetrics creates new user service metrics instance
func NewUserMetrics() UserMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("user-service", nil)

	userMetrics := &UserPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	// Initialize gRPC metrics
	userMetrics.grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "user_service",
			Name:      "grpc_request_duration_seconds",
			Help:      "gRPC request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	userMetrics.grpcRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "user_service",
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	// Register gRPC metrics
	baseMetrics.GetRegistry().MustRegister(
		userMetrics.grpcRequestDuration,
		userMetrics.grpcRequestTotal,
	)

	return userMetrics
}

// GRPCRequestDuration records gRPC request duration
func (m *UserPrometheusMetrics) GRPCRequestDuration(method string, duration time.Duration) {
	m.grpcRequestDuration.WithLabelValues("user-service", method).Observe(duration.Seconds())
}

// GRPCRequestTotal increments gRPC request counter
func (m *UserPrometheusMetrics) GRPCRequestTotal(method, status string) {
	m.grpcRequestTotal.WithLabelValues("user-service", method, status).Inc()
}
