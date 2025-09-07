package outbox

import (
	"context"
	"time"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/publisher"
)

// BackgroundPublisher handles background publishing of outbox events
type BackgroundPublisher struct {
	repo      outbox.Repository
	publisher publisher.StockEventsPublisher
	logger    logger.Logger

	interval  time.Duration
	batchSize int
}

// NewBackgroundPublisher creates a new publisher
func NewBackgroundPublisher(repo outbox.Repository, stockPublisher publisher.StockEventsPublisher, logger logger.Logger, interval time.Duration, batchSize int) *BackgroundPublisher {
	return &BackgroundPublisher{
		repo:      repo,
		publisher: stockPublisher,
		logger:    logger,
		interval:  interval,
		batchSize: batchSize,
	}
}

// Start runs the publisher loop until the context is canceled
func (bp *BackgroundPublisher) Start(ctx context.Context) {
	ticker := time.NewTicker(bp.interval)
	defer ticker.Stop()

	bp.logger.Info("starting background outbox publisher", "interval", bp.interval, "batchSize", bp.batchSize)

	for {
		select {
		case <-ctx.Done():
			bp.logger.Info("background outbox publisher stopped", "reason", ctx.Err())
			return
		case <-ticker.C:
			if err := bp.processBatch(ctx); err != nil {
				bp.logger.Error("failed to process batch", "error", err)
			}
		}
	}
}

// processBatch fetches unprocessed events and publishes them
func (bp *BackgroundPublisher) processBatch(ctx context.Context) error {
	events, err := bp.repo.GetUnprocessedEvents(ctx, bp.batchSize)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	bp.logger.Info("processing outbox events", "count", len(events))

	for _, e := range events {
		// Publish using direct publisher
		if err := bp.publisher.PublishFromOutbox(ctx, e); err != nil {
			bp.logger.Error("failed to publish event", "eventID", e.ID, "eventType", e.Type, "err", err)
			_ = bp.repo.MarkAsFailed(ctx, e.ID, err.Error())
			continue
		}

		// Mark processed
		if err := bp.repo.MarkAsProcessed(ctx, e.ID); err != nil {
			bp.logger.Error("failed to mark event as processed", "eventID", e.ID, "err", err)
		}
	}

	return nil
}
