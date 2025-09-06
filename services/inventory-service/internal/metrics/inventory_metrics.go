package metrics

import (
	"ecommerce-platform/pkg/metrics"
)

// InventoryMetrics interface defines inventory service specific metrics
type InventoryMetrics interface {
	// Product business metrics
	ProductCreated()
	ProductCreationFailed(reason string)

	// Stock management metrics
	StockReserved()
	StockReservationFailed(reason string)
	StockReleased()
	StockCommitted()

	// HTTP metrics (reused from pkg/metrics)
	metrics.Metrics
}

// InventoryPrometheusMetrics implements InventoryMetrics interface
type InventoryPrometheusMetrics struct {
	*metrics.PrometheusMetrics
}

// NewInventoryMetrics creates new inventory service metrics instance
func NewInventoryMetrics() InventoryMetrics {
	baseMetrics := metrics.NewPrometheusMetrics("inventory-service", nil)

	inventoryMetrics := &InventoryPrometheusMetrics{
		PrometheusMetrics: baseMetrics,
	}

	return inventoryMetrics
}

// ProductCreated increments product creation counter
func (m *InventoryPrometheusMetrics) ProductCreated() {
	m.EntityEvent(metrics.EntityTypeProduct, metrics.ActionCreated, "")
}

// ProductCreationFailed increments product creation failure counter with reason
func (m *InventoryPrometheusMetrics) ProductCreationFailed(reason string) {
	m.EntityEvent(metrics.EntityTypeProduct, metrics.ActionFailed, reason)
}

// StockReserved increments stock reservation counter
func (m *InventoryPrometheusMetrics) StockReserved() {
	m.EntityEvent(metrics.EntityTypeProduct, metrics.ActionStockReserved, "")
}

// StockReservationFailed increments stock reservation failure counter with reason
func (m *InventoryPrometheusMetrics) StockReservationFailed(reason string) {
	m.EntityEvent(metrics.EntityTypeProduct, metrics.ActionFailed, reason)
}

// StockReleased increments stock release counter
func (m *InventoryPrometheusMetrics) StockReleased() {
	m.EntityEvent(metrics.EntityTypeProduct, metrics.ActionStockReleased, "")
}

// StockCommitted increments stock commit counter
func (m *InventoryPrometheusMetrics) StockCommitted() {
	m.EntityEvent(metrics.EntityTypeProduct, metrics.ActionStockCommitted, "")
}
