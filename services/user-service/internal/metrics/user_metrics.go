package metrics

import (
	"github.com/kubernetestest/ecommerce-platform/pkg/metrics"
)

// UserMetrics interface defines user service specific metrics
type UserMetrics interface {
	// User business metrics
	UserCreated()
	UserLoginSuccess()
	UserLoginFailed(reason string)
	UserLogout()
	UserProfileUpdated()

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// UserPrometheusMetrics implements UserMetrics interface
type UserPrometheusMetrics struct {
	*metrics.PrometheusMetrics
}

// NewUserMetrics creates new user service metrics instance
func NewUserMetrics() UserMetrics {
	return &UserPrometheusMetrics{
		PrometheusMetrics: metrics.NewPrometheusMetrics("user-service"),
	}
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

// UserLogout increments logout counter
func (m *UserPrometheusMetrics) UserLogout() {
	m.EntityEvent(metrics.EntityTypeUser, metrics.ActionLogout, "")
}

// UserProfileUpdated increments profile update counter
func (m *UserPrometheusMetrics) UserProfileUpdated() {
	m.EntityEvent(metrics.EntityTypeUser, metrics.ActionUpdated, "")
}
