package consumer

import (
	"context"
	"fmt"
	"strings"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/payment-service/internal/application/services"
	"ecommerce-platform/services/payment-service/internal/domain/ports/consumer"
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

// StartStockReservedConsumer starts the stock reserved consumer
func (cm *ConsumerManager) StartStockReservedConsumer(
	ctx context.Context,
	config consumer.ConsumerConfig,
	applicationService interface{},
) error {
	appSvc, ok := applicationService.(*services.PaymentApplicationService)
	if !ok {
		return fmt.Errorf("invalid application service type")
	}
	handlers := NewEventHandlers(appSvc, cm.logger)

	consumer, err := NewEventConsumer[*events.StockReserved](
		strings.Join(config.BootstrapServers, ","),
		config.GroupID,
		config.AutoOffsetReset,
		EventHandlerFunc[*events.StockReserved](handlers.HandleStockReserved),
	)
	if err != nil {
		return fmt.Errorf("failed to create stock reserved consumer: %w", err)
	}

	cm.consumers = append(cm.consumers, consumer)

	// Start consumer in background
	go func() {
		cm.logger.Info("starting stock reserved consumer", "topics", config.Topics)
		if err := consumer.Run(ctx, config.Topics); err != nil {
			cm.logger.Error("stock reserved consumer failed", "error", err)
		}
	}()

	cm.logger.Info("stock reserved consumer started")
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
