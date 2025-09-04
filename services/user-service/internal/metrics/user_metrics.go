package metrics

import (
	"ecommerce-platform/pkg/metrics"
)

// UserMetrics interface defines user service specific metrics
type UserMetrics interface {
	// User business metrics
	UserCreated()
	UserLoginSuccess()
	UserLoginFailed(reason string)

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// UserPrometheusMetrics implements UserMetrics interface
type UserPrometheusMetrics struct {
	*metrics.PrometheusMetrics
}

// NewUserMetrics creates new user service metrics instance
func NewUserMetrics() UserMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("user-service", nil)

	userMetrics := &UserPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	// Note: User service uses only base metrics from pkg/metrics
	// No additional service-specific metrics needed for now

	return userMetrics
}

// UserCreated increments user creation counter
func (m *UserPrometheusMetrics) UserCreated() {
	m.EntityEvent(metrics.EntityTypeUser, metrics.ActionCreated, "")
}

// UserLoginSuccess increments successful login counter
func (m *UserPrometheusMetrics) UserLoginSuccess() {
	m.EntityEvent(metrics.EntityTypeUser, metrics.ActionLoginSuccess, "")
}

// UserLoginFailed increments failed login counter with reason
func (m *UserPrometheusMetrics) UserLoginFailed(reason string) {
	m.EntityEvent(metrics.EntityTypeUser, metrics.ActionLoginFailed, reason)
}
