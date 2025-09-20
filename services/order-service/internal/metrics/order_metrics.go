package metrics

import (
	"time"

	"ecommerce-platform/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// OrderMetrics interface defines order service specific metrics
type OrderMetrics interface {
	// gRPC metrics
	GRPCRequestDuration(method string, duration time.Duration)
	GRPCRequestTotal(method, status string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// OrderPrometheusMetrics implements OrderMetrics interface
type OrderPrometheusMetrics struct {
	*metrics.PrometheusMetrics

	// gRPC metrics
	grpcRequestDuration *prometheus.HistogramVec
	grpcRequestTotal    *prometheus.CounterVec
}

// NewOrderMetrics creates new order service metrics instance
func NewOrderMetrics() OrderMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("order-service", nil)

	orderMetrics := &OrderPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	// Initialize gRPC metrics
	orderMetrics.grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "order_service",
			Name:      "grpc_request_duration_seconds",
			Help:      "gRPC request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	orderMetrics.grpcRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	// Register gRPC metrics
	baseMetrics.GetRegistry().MustRegister(
		orderMetrics.grpcRequestDuration,
		orderMetrics.grpcRequestTotal,
	)

	return orderMetrics
}

// GRPCRequestDuration records gRPC request duration
func (m *OrderPrometheusMetrics) GRPCRequestDuration(method string, duration time.Duration) {
	m.grpcRequestDuration.WithLabelValues("order-service", method).Observe(duration.Seconds())
}

// GRPCRequestTotal increments gRPC request counter
func (m *OrderPrometheusMetrics) GRPCRequestTotal(method, status string) {
	m.grpcRequestTotal.WithLabelValues("order-service", method, status).Inc()
}
