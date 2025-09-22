package metrics

import (
	"time"

	"ecommerce-platform/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// PaymentMetrics interface defines payment service specific metrics
type PaymentMetrics interface {
	// gRPC metrics
	GRPCRequestDuration(method string, duration time.Duration)
	GRPCRequestTotal(method, status string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// PaymentPrometheusMetrics implements PaymentMetrics interface
type PaymentPrometheusMetrics struct {
	*metrics.PrometheusMetrics

	// gRPC metrics
	grpcRequestDuration *prometheus.HistogramVec
	grpcRequestTotal    *prometheus.CounterVec
}

// NewPaymentMetrics creates new payment service metrics instance
func NewPaymentMetrics() PaymentMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("payment-service", nil)

	paymentMetrics := &PaymentPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	// Initialize gRPC metrics
	paymentMetrics.grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "payment_service",
			Name:      "grpc_request_duration_seconds",
			Help:      "gRPC request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	paymentMetrics.grpcRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "payment_service",
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	// Register gRPC metrics
	baseMetrics.GetRegistry().MustRegister(
		paymentMetrics.grpcRequestDuration,
		paymentMetrics.grpcRequestTotal,
	)

	return paymentMetrics
}

// GRPCRequestDuration records gRPC request duration
func (m *PaymentPrometheusMetrics) GRPCRequestDuration(method string, duration time.Duration) {
	m.grpcRequestDuration.WithLabelValues("payment-service", method).Observe(duration.Seconds())
}

// GRPCRequestTotal increments gRPC request counter
func (m *PaymentPrometheusMetrics) GRPCRequestTotal(method, status string) {
	m.grpcRequestTotal.WithLabelValues("payment-service", method, status).Inc()
}
