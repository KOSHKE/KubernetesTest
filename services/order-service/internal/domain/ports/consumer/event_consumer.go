package consumer

import (
	"context"
)

// EventConsumerManager defines the interface for managing event consumers
type EventConsumerManager interface {
	// StartStockConsumer starts all stock-related event consumers
	StartStockConsumer(ctx context.Context, config ConsumerConfig, applicationService interface{}) error

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
