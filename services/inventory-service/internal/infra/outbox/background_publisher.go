package outbox

import (
	"context"
	"sync"
	"time"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// BackgroundPublisher handles background publishing of outbox events using hybrid approach
type BackgroundPublisher struct {
	// Dependencies
	outboxRepo repository.OutboxRepository
	publisher  outbox.Publisher
	logger     logger.Logger

	// Configuration
	workers    int           // Number of worker goroutines
	interval   time.Duration // How often to check for new events
	batchSize  int           // Maximum events to process per batch
	maxRetries int           // Maximum retry attempts for failed events
	retryDelay time.Duration // Delay between retries

	// Internal channels
	jobQueue   chan outbox.Event
	errorQueue chan EventError

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// EventError represents an error that occurred while processing an event
type EventError struct {
	Event outbox.Event
	Error error
	Retry int
}

// NewBackgroundPublisher creates a new background publisher
func NewBackgroundPublisher(
	outboxRepo repository.OutboxRepository,
	publisher outbox.Publisher,
	logger logger.Logger,
) *BackgroundPublisher {
	return &BackgroundPublisher{
		outboxRepo: outboxRepo,
		publisher:  publisher,
		logger:     logger,
		interval:   5 * time.Second, // Check every 5 seconds
		batchSize:  100,             // Process up to 100 events per batch
	}
}

// Start begins the background publishing process
func (bp *BackgroundPublisher) Start(ctx context.Context) {
	bp.logger.Info("starting background outbox publisher", "interval", bp.interval, "batchSize", bp.batchSize)

	ticker := time.NewTicker(bp.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			bp.logger.Info("background outbox publisher stopped", "reason", ctx.Err())
			return
		case <-ticker.C:
			if err := bp.processOutboxEvents(ctx); err != nil {
				bp.logger.Error("failed to process outbox events", "error", err)
			}
		}
	}
}

// processOutboxEvents processes unprocessed events from outbox
func (bp *BackgroundPublisher) processOutboxEvents(ctx context.Context) error {
	// Get unprocessed events
	events, err := bp.outboxRepo.GetUnprocessedEvents(ctx, bp.batchSize)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil // No events to process
	}

	bp.logger.Info("processing outbox events", "count", len(events))

	// Process events in batches
	for _, event := range events {
		if err := bp.publishEvent(ctx, event); err != nil {
			bp.logger.Error("failed to publish event", "eventID", event.ID, "error", err)
			continue // Continue with next event
		}

		// Mark event as processed
		if err := bp.outboxRepo.MarkAsProcessed(ctx, event.ID); err != nil {
			bp.logger.Error("failed to mark event as processed", "eventID", event.ID, "error", err)
		}
	}

	bp.logger.Info("outbox events processed successfully", "count", len(events))
	return nil
}

// publishEvent publishes a single event
func (bp *BackgroundPublisher) publishEvent(ctx context.Context, event outbox.Event) error {
	// Publish event using the publisher
	if err := bp.publisher.Publish(ctx, event); err != nil {
		return err
	}

	bp.logger.Debug("event published successfully", "eventID", event.ID, "type", event.Type)
	return nil
}
