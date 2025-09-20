package consumer

import (
	"context"
	"fmt"
	"strings"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/order-service/internal/application/services"
	"ecommerce-platform/services/order-service/internal/domain/ports/consumer"
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

// StartStockConsumer starts the stock events consumer
func (cm *ConsumerManager) StartStockConsumer(
	ctx context.Context,
	config consumer.ConsumerConfig,
	applicationService interface{},
) error {
	appSvc, ok := applicationService.(*services.OrderApplicationService)
	if !ok {
		return fmt.Errorf("invalid application service type")
	}
	handlers := NewEventHandlers(appSvc, cm.logger)

	// Start StockReserved consumer
	stockReservedConsumer, err := NewEventConsumer[*events.StockReserved](
		strings.Join(config.BootstrapServers, ","),
		config.GroupID+"-stock-reserved",
		config.AutoOffsetReset,
		EventHandlerFunc[*events.StockReserved](handlers.HandleStockReserved),
	)
	if err != nil {
		return fmt.Errorf("failed to create stock reserved consumer: %w", err)
	}
	cm.consumers = append(cm.consumers, stockReservedConsumer)

	// Note: StockReservationFailed event doesn't exist in proto
	// Removed StockReservationFailed consumer

	// Start StockReleased consumer
	stockReleasedConsumer, err := NewEventConsumer[*events.StockReleased](
		strings.Join(config.BootstrapServers, ","),
		config.GroupID+"-stock-released",
		config.AutoOffsetReset,
		EventHandlerFunc[*events.StockReleased](handlers.HandleStockReleased),
	)
	if err != nil {
		return fmt.Errorf("failed to create stock released consumer: %w", err)
	}
	cm.consumers = append(cm.consumers, stockReleasedConsumer)

	// Start StockCommitted consumer
	stockCommittedConsumer, err := NewEventConsumer[*events.StockCommitted](
		strings.Join(config.BootstrapServers, ","),
		config.GroupID+"-stock-committed",
		config.AutoOffsetReset,
		EventHandlerFunc[*events.StockCommitted](handlers.HandleStockCommitted),
	)
	if err != nil {
		return fmt.Errorf("failed to create stock committed consumer: %w", err)
	}
	cm.consumers = append(cm.consumers, stockCommittedConsumer)

	// Start all stock consumers in background
	stockTopics := []string{"stock-events"} // Assuming all stock events are in one topic

	go func() {
		cm.logger.Info("starting stock reserved consumer", "topics", stockTopics)
		if err := stockReservedConsumer.Run(ctx, stockTopics); err != nil {
			cm.logger.Error("stock reserved consumer failed", "error", err)
		}
	}()

	// Removed StockReservationFailed consumer goroutine

	go func() {
		cm.logger.Info("starting stock released consumer", "topics", stockTopics)
		if err := stockReleasedConsumer.Run(ctx, stockTopics); err != nil {
			cm.logger.Error("stock released consumer failed", "error", err)
		}
	}()

	go func() {
		cm.logger.Info("starting stock committed consumer", "topics", stockTopics)
		if err := stockCommittedConsumer.Run(ctx, stockTopics); err != nil {
			cm.logger.Error("stock committed consumer failed", "error", err)
		}
	}()

	cm.logger.Info("stock consumers started")
	return nil
}

// StartPaymentConsumer starts the payment processed consumer
func (cm *ConsumerManager) StartPaymentConsumer(
	ctx context.Context,
	config consumer.ConsumerConfig,
	applicationService interface{},
) error {
	appSvc, ok := applicationService.(*services.OrderApplicationService)
	if !ok {
		return fmt.Errorf("invalid application service type")
	}
	handlers := NewEventHandlers(appSvc, cm.logger)

	consumer, err := NewEventConsumer[*events.PaymentProcessed](
		strings.Join(config.BootstrapServers, ","),
		config.GroupID+"-payment-processed",
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
