package metrics

import (
	"ecommerce-platform/pkg/metrics"
)

// OrderMetrics interface defines order service specific metrics
type OrderMetrics interface {
	// Order business metrics
	OrderCreated(currency string)
	OrderCreationFailed(reason string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// OrderPrometheusMetrics implements OrderMetrics interface
type OrderPrometheusMetrics struct {
	*metrics.PrometheusMetrics
}

// NewOrderMetrics creates new order service metrics instance
func NewOrderMetrics() OrderMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("order-service", nil)

	orderMetrics := &OrderPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	// Note: Order service uses only base metrics from pkg/metrics
	// No additional service-specific metrics needed for now

	return orderMetrics
}

// OrderCreated increments order creation counter
func (m *OrderPrometheusMetrics) OrderCreated(currency string) {
	// Use base EntityEvent with order entity type and created action
	m.EntityEvent(metrics.EntityTypeOrder, metrics.ActionCreated, "")
}

// OrderCreationFailed increments order creation failure counter
func (m *OrderPrometheusMetrics) OrderCreationFailed(reason string) {
	if reason == "" {
		reason = "unknown"
	}
	m.EntityEvent(metrics.EntityTypeOrder, metrics.ActionFailed, reason)
}
