package metrics

import (
	"time"

	"github.com/kubernetestest/ecommerce-platform/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// OrderMetrics interface defines order service specific metrics
type OrderMetrics interface {
	// Order business metrics
	OrderCreated(currency string)
	OrderCreationFailed(reason string)

	// Event processing metrics
	EventProcessed(eventType string, success bool)
	EventProcessingDuration(duration time.Duration, eventType string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// OrderPrometheusMetrics implements OrderMetrics interface
type OrderPrometheusMetrics struct {
	*metrics.PrometheusMetrics

	// Business metrics with consistent labels (matching pkg/metrics)
	orderCreatedTotal        *prometheus.CounterVec
	orderCreationFailedTotal *prometheus.CounterVec

	// Event processing metrics
	eventProcessedTotal *prometheus.CounterVec
	eventDuration       *prometheus.HistogramVec
}

// NewOrderMetrics creates new order service metrics instance
func NewOrderMetrics() OrderMetrics {
	return &OrderPrometheusMetrics{
		PrometheusMetrics: metrics.NewPrometheusMetrics("order-service"),

		// Business metrics with consistent labels
		orderCreatedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_created_total",
				Help: "Total number of orders created",
			},
			[]string{"service", "currency", "status"},
		),
		orderCreationFailedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_creation_failed_total",
				Help: "Total number of failed order creation attempts",
			},
			[]string{"service", "reason"},
		),

		// Event processing metrics
		eventProcessedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "event_processed_total",
				Help: "Total number of events processed",
			},
			[]string{"service", "event_type", "status"},
		),
		eventDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "event_processing_duration_seconds",
				Help:    "Event processing duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "event_type"},
		),
	}
}

// OrderCreated increments order creation counter
func (m *OrderPrometheusMetrics) OrderCreated(currency string) {
	m.orderCreatedTotal.WithLabelValues("order-service", currency, "success").Inc()
}

// OrderCreationFailed increments order creation failure counter
func (m *OrderPrometheusMetrics) OrderCreationFailed(reason string) {
	if reason == "" {
		reason = "unknown"
	}
	m.orderCreationFailedTotal.WithLabelValues("order-service", reason).Inc()
}

// EventProcessed increments event processing counter
func (m *OrderPrometheusMetrics) EventProcessed(eventType string, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	m.eventProcessedTotal.WithLabelValues("order-service", eventType, status).Inc()
}

// EventProcessingDuration records event processing time
func (m *OrderPrometheusMetrics) EventProcessingDuration(duration time.Duration, eventType string) {
	m.eventDuration.WithLabelValues("order-service", eventType).Observe(duration.Seconds())
}
