package consumer

import (
	"context"
	"ecommerce-platform/proto-go/events"
)

// EventConsumer defines the interface for consuming events
// This interface is implemented by infrastructure layer
type EventConsumer interface {
	// HandleOrderCreated handles when a new order is created
	HandleOrderCreated(ctx context.Context, event *events.OrderCreated) error

	// HandlePaymentProcessed handles when payment is processed
	HandlePaymentProcessed(ctx context.Context, event *events.PaymentProcessed) error

	// Close closes the consumer
	Close() error
}

// EventConsumerManager defines the interface for managing event consumers
type EventConsumerManager interface {
	// StartOrderConsumer starts the order created consumer
	StartOrderConsumer(ctx context.Context, config ConsumerConfig, applicationService interface{}) error

	// StartPaymentConsumer starts the payment processed consumer
	StartPaymentConsumer(ctx context.Context, config ConsumerConfig, applicationService interface{}) error

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
