package consumer

import (
	"context"
	"ecommerce-platform/proto-go/events"
)

// EventConsumer defines the interface for consuming events
// This interface is implemented by infrastructure layer
type EventConsumer interface {
	// HandleStockReserved handles when stock is reserved for an order
	HandleStockReserved(ctx context.Context, event *events.StockReserved) error

	// Close closes the consumer
	Close() error
}

// EventConsumerManager defines the interface for managing event consumers
type EventConsumerManager interface {
	// StartStockReservedConsumer starts the stock reserved consumer
	StartStockReservedConsumer(ctx context.Context, config ConsumerConfig, applicationService interface{}) error

	// Close closes all consumers
	Close() error
}

// ConsumerConfig holds configuration for a consumer
type ConsumerConfig struct {
	BootstrapServers []string
	GroupID          string
	AutoOffsetReset  string
	Topics           []string
}
