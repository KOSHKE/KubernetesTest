package health

import (
	"context"
	"sync"
	"time"
)

// Checker defines the interface for health checks
type Checker interface {
	// Name returns the name of the component being checked
	Name() string

	// Check performs the health check and returns an error if unhealthy
	Check(ctx context.Context) error
}

// Manager manages multiple health checkers
type Manager struct {
	checkers []Checker
	mu       sync.RWMutex
}

// NewManager creates a new health manager
func NewManager() *Manager {
	return &Manager{
		checkers: make([]Checker, 0),
	}
}

// AddChecker adds a health checker
func (m *Manager) AddChecker(checker Checker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers = append(m.checkers, checker)
}

// IsHealthy checks if all components are healthy
func (m *Manager) IsHealthy(ctx context.Context) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, checker := range m.checkers {
		if err := checker.Check(ctx); err != nil {
			return false
		}
	}
	return true
}

// IsReady is an alias for IsHealthy (for readiness checks)
func (m *Manager) IsReady(ctx context.Context) bool {
	return m.IsHealthy(ctx)
}

// GetStatus returns the status of all components
func (m *Manager) GetStatus(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]error)
	for _, checker := range m.checkers {
		// Use timeout for each check
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := checker.Check(checkCtx)
		cancel()

		status[checker.Name()] = err
	}
	return status
}
