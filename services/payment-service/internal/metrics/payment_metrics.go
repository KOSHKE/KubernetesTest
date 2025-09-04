package metrics

import (
	"ecommerce-platform/pkg/metrics"
)

// PaymentMetrics interface defines payment service specific metrics
type PaymentMetrics interface {
	// Payment business metrics
	PaymentSucceeded(method string)
	PaymentFailed(reason string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// PaymentPrometheusMetrics implements PaymentMetrics interface
type PaymentPrometheusMetrics struct {
	*metrics.PrometheusMetrics
}

// NewPaymentMetrics creates new payment service metrics instance
func NewPaymentMetrics() PaymentMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("payment-service", nil)

	paymentMetrics := &PaymentPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	// Note: Payment service uses only base metrics from pkg/metrics
	// No additional service-specific metrics needed for now

	return paymentMetrics
}

// PaymentSucceeded increments successful payment counter
func (m *PaymentPrometheusMetrics) PaymentSucceeded(method string) {
	// Use base EntityEvent with payment entity type and succeeded action
	m.EntityEvent(metrics.EntityTypePayment, metrics.ActionSucceeded, "")
}

// PaymentFailed increments failed payment counter with reason
func (m *PaymentPrometheusMetrics) PaymentFailed(reason string) {
	if reason == "" {
		reason = "unknown"
	}
	m.EntityEvent(metrics.EntityTypePayment, metrics.ActionFailed, reason)
}
