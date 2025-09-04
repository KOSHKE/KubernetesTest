package metrics

import (
	"ecommerce-platform/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// InventoryMetrics defines inventory-specific metrics
type InventoryMetrics struct {
	*metrics.PrometheusMetrics
}

// NewInventoryMetrics creates new inventory metrics
func NewInventoryMetrics() *InventoryMetrics {
	return &InventoryMetrics{
		PrometheusMetrics: metrics.NewPrometheusMetrics("inventory_service", prometheus.DefaultRegisterer),
	}
}

// RecordProductCreated records a product creation
func (m *InventoryMetrics) RecordProductCreated() {
	m.RecordHTTPRequest("POST", "/products", 200)
}

// RecordProductCreationFailed records a product creation failure
func (m *InventoryMetrics) RecordProductCreationFailed() {
	m.RecordHTTPRequest("POST", "/products", 500)
}

// RecordStockReserved records a stock reservation
func (m *InventoryMetrics) RecordStockReserved() {
	m.RecordHTTPRequest("POST", "/stock/reserve", 200)
}

// RecordStockReservationFailed records a stock reservation failure
func (m *InventoryMetrics) RecordStockReservationFailed() {
	m.RecordHTTPRequest("POST", "/stock/reserve", 500)
}

// RecordStockReleased records a stock release
func (m *InventoryMetrics) RecordStockReleased() {
	m.RecordHTTPRequest("POST", "/stock/release", 200)
}

// RecordStockCommitted records a stock commit
func (m *InventoryMetrics) RecordStockCommitted() {
	m.RecordHTTPRequest("POST", "/stock/commit", 200)
}
