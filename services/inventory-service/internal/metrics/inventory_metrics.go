package metrics

import (
	"time"

	"ecommerce-platform/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// InventoryMetrics interface defines inventory service specific metrics
type InventoryMetrics interface {
	// gRPC metrics
	GRPCRequestDuration(method string, duration time.Duration)
	GRPCRequestTotal(method, status string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// InventoryPrometheusMetrics implements InventoryMetrics interface
type InventoryPrometheusMetrics struct {
	*metrics.PrometheusMetrics

	// gRPC metrics
	grpcRequestDuration *prometheus.HistogramVec
	grpcRequestTotal    *prometheus.CounterVec
}

// NewInventoryMetrics creates new inventory service metrics instance
func NewInventoryMetrics() InventoryMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("inventory-service", nil)

	inventoryMetrics := &InventoryPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	// Initialize gRPC metrics
	inventoryMetrics.grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "inventory_service",
			Name:      "grpc_request_duration_seconds",
			Help:      "gRPC request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	inventoryMetrics.grpcRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "inventory_service",
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	// Register gRPC metrics
	baseMetrics.GetRegistry().MustRegister(
		inventoryMetrics.grpcRequestDuration,
		inventoryMetrics.grpcRequestTotal,
	)

	return inventoryMetrics
}

// GRPCRequestDuration records gRPC request duration
func (m *InventoryPrometheusMetrics) GRPCRequestDuration(method string, duration time.Duration) {
	m.grpcRequestDuration.WithLabelValues("inventory-service", method).Observe(duration.Seconds())
}

// GRPCRequestTotal increments gRPC request counter
func (m *InventoryPrometheusMetrics) GRPCRequestTotal(method, status string) {
	m.grpcRequestTotal.WithLabelValues("inventory-service", method, status).Inc()
}
