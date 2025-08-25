package metrics

import (
	"time"

	"github.com/kubernetestest/ecommerce-platform/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PaymentMetrics interface defines payment service specific metrics
type PaymentMetrics interface {
	// Payment business metrics
	PaymentSucceeded(method string)
	PaymentFailed(reason string)

	// Processing time metrics
	PaymentProcessingDuration(duration time.Duration, method string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// PaymentPrometheusMetrics implements PaymentMetrics interface
type PaymentPrometheusMetrics struct {
	*metrics.PrometheusMetrics

	// Business metrics
	paymentSucceededTotal *prometheus.CounterVec
	paymentFailedTotal    *prometheus.CounterVec
	paymentDuration       *prometheus.HistogramVec
}

// NewPaymentMetrics creates new payment service metrics instance
func NewPaymentMetrics() PaymentMetrics {
	return &PaymentPrometheusMetrics{
		PrometheusMetrics: metrics.NewPrometheusMetrics("payment-service"),

		// Business metrics with consistent labels (matching pkg/metrics)
		paymentSucceededTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_succeeded_total",
				Help: "Total number of successful payments",
			},
			[]string{"service", "method", "status"},
		),
		paymentFailedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_failed_total",
				Help: "Total number of failed payments",
			},
			[]string{"service", "method", "failure_reason"},
		),
		paymentDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_processing_duration_seconds",
				Help:    "Payment processing duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),
	}
}

// PaymentSucceeded increments successful payment counter
func (m *PaymentPrometheusMetrics) PaymentSucceeded(method string) {
	m.paymentSucceededTotal.WithLabelValues("payment-service", method, "success").Inc()
}

// PaymentFailed increments failed payment counter with reason
func (m *PaymentPrometheusMetrics) PaymentFailed(reason string) {
	m.paymentFailedTotal.WithLabelValues("payment-service", "unknown", reason).Inc()
}

// PaymentProcessingDuration records payment processing time
func (m *PaymentPrometheusMetrics) PaymentProcessingDuration(duration time.Duration, method string) {
	m.paymentDuration.WithLabelValues("payment-service", method).Observe(duration.Seconds())
}
