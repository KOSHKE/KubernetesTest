package consumer

import (
	"context"
	"fmt"
	"strings"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/consumer"
)

// ConsumerManager manages all Kafka consumers
// Implements consumer.EventConsumerManager interface
type ConsumerManager struct {
	consumers []interface{ Close() error }
	logger    logger.Logger
}

// NewConsumerManager creates a new ConsumerManager
func NewConsumerManager(logger logger.Logger) consumer.EventConsumerManager {
	return &ConsumerManager{
		consumers: make([]interface{ Close() error }, 0),
		logger:    logger,
	}
}

// StartOrderConsumer starts the order created consumer
func (cm *ConsumerManager) StartOrderConsumer(
	ctx context.Context,
	config consumer.ConsumerConfig,
	applicationService interface{},
) error {
	appSvc, ok := applicationService.(*services.InventoryApplicationService)
	if !ok {
		return fmt.Errorf("invalid application service type")
	}
	handlers := NewEventHandlers(appSvc, cm.logger)

	consumer, err := NewEventConsumer[*events.OrderCreated](
		strings.Join(config.BootstrapServers, ","),
		config.GroupID,
		config.AutoOffsetReset,
		EventHandlerFunc[*events.OrderCreated](handlers.HandleOrderCreated),
	)
	if err != nil {
		return fmt.Errorf("failed to create order consumer: %w", err)
	}

	cm.consumers = append(cm.consumers, consumer)

	// Start consumer in background
	go func() {
		cm.logger.Info("starting order consumer", "topics", config.Topics)
		if err := consumer.Run(ctx, config.Topics); err != nil {
			cm.logger.Error("order consumer failed", "error", err)
		}
	}()

	cm.logger.Info("order consumer started")
	return nil
}

// StartPaymentConsumer starts the payment processed consumer
func (cm *ConsumerManager) StartPaymentConsumer(
	ctx context.Context,
	config consumer.ConsumerConfig,
	applicationService interface{},
) error {
	appSvc, ok := applicationService.(*services.InventoryApplicationService)
	if !ok {
		return fmt.Errorf("invalid application service type")
	}
	handlers := NewEventHandlers(appSvc, cm.logger)

	consumer, err := NewEventConsumer[*events.PaymentProcessed](
		strings.Join(config.BootstrapServers, ","),
		config.GroupID,
		config.AutoOffsetReset,
		EventHandlerFunc[*events.PaymentProcessed](handlers.HandlePaymentProcessed),
	)
	if err != nil {
		return fmt.Errorf("failed to create payment consumer: %w", err)
	}

	cm.consumers = append(cm.consumers, consumer)

	// Start consumer in background
	go func() {
		cm.logger.Info("starting payment consumer", "topics", config.Topics)
		if err := consumer.Run(ctx, config.Topics); err != nil {
			cm.logger.Error("payment consumer failed", "error", err)
		}
	}()

	cm.logger.Info("payment consumer started")
	return nil
}

// Close closes all consumers
func (cm *ConsumerManager) Close() error {
	var firstErr error
	for i, consumer := range cm.consumers {
		if err := consumer.Close(); err != nil {
			cm.logger.Error("failed to close consumer", "index", i, "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
