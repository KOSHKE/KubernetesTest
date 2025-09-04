package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Entity action constants
const (
	ActionCreated      = "created"
	ActionUpdated      = "updated"
	ActionDeleted      = "deleted"
	ActionLoginSuccess = "login_success"
	ActionLoginFailed  = "login_failed"
	ActionLogout       = "logout"
	ActionSucceeded    = "succeeded"
	ActionFailed       = "failed"
)

// Common reason constants
const (
	ReasonNone = "none"
)

// Entity type constants
const (
	EntityTypeUser    = "user"
	EntityTypeOrder   = "order"
	EntityTypePayment = "payment"
	EntityTypeProduct = "product"
)

// Metrics interface defines all available metrics
type Metrics interface {
	// Entity metrics
	EntityEvent(entityType, action, reason string)

	// HTTP metrics
	HTTPRequestsTotal(method, endpoint, status string)
	HTTPRequestDuration(method, endpoint string, duration time.Duration)
}

// PrometheusMetrics implements Metrics interface using Prometheus
type PrometheusMetrics struct {
	serviceName string
	registry    prometheus.Registerer

	// Entity events counter
	entityEventsCounter *prometheus.CounterVec

	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
}

// NewPrometheusMetrics creates new Prometheus metrics instance
func NewPrometheusMetrics(serviceName string, registry prometheus.Registerer) *PrometheusMetrics {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}

	metrics := &PrometheusMetrics{
		serviceName: serviceName,
		registry:    registry,
	}

	// Entity events counter
	metrics.entityEventsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "entity_events_total",
			Help:      "Total number of entity events",
		},
		[]string{"service", "entity_type", "action", "reason"},
	)

	// HTTP metrics
	metrics.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	metrics.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: serviceName,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets, // Standard Prometheus buckets
		},
		[]string{"service", "method", "endpoint"},
	)

	// Register all metrics with the registry
	registry.MustRegister(
		metrics.entityEventsCounter,
		metrics.httpRequestsTotal,
		metrics.httpRequestDuration,
	)

	return metrics
}

// EntityEvent increments entity event counter
func (m *PrometheusMetrics) EntityEvent(entityType, action, reason string) {
	// Replace empty reason with "none" to avoid empty labels in Prometheus
	if reason == "" {
		reason = ReasonNone
	}
	m.entityEventsCounter.WithLabelValues(m.serviceName, entityType, action, reason).Inc()
}

// HTTPRequestsTotal increments HTTP request counter
func (m *PrometheusMetrics) HTTPRequestsTotal(method, endpoint, status string) {
	m.httpRequestsTotal.WithLabelValues(m.serviceName, method, endpoint, status).Inc()
}

// HTTPRequestDuration records HTTP request duration
func (m *PrometheusMetrics) HTTPRequestDuration(method, endpoint string, duration time.Duration) {
	m.httpRequestDuration.WithLabelValues(m.serviceName, method, endpoint).Observe(duration.Seconds())
}

// GetRegistry returns the underlying prometheus registry
func (m *PrometheusMetrics) GetRegistry() prometheus.Registerer {
	return m.registry
}

// ResetMetrics removes all metrics from the registry (mainly for testing)
func (m *PrometheusMetrics) ResetMetrics() {
	m.registry.Unregister(m.entityEventsCounter)
	m.registry.Unregister(m.httpRequestsTotal)
	m.registry.Unregister(m.httpRequestDuration)
}
