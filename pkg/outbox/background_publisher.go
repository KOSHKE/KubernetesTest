package outbox

import (
	"context"
	"time"

	"ecommerce-platform/pkg/logger"
)

// EventPublisher defines interface for publishing events from outbox
// This allows different services to provide their own publisher implementations
type EventPublisher interface {
	PublishFromOutbox(ctx context.Context, event Event) error
}

// BackgroundPublisher handles background publishing of outbox events
// This is a generic implementation that can be used across all services
type BackgroundPublisher struct {
	repo      Repository
	publisher EventPublisher
	logger    logger.Logger

	interval  time.Duration
	batchSize int

	// shutdown control
	cancel context.CancelFunc
	done   chan struct{}
}

// NewBackgroundPublisher creates a new background publisher
func NewBackgroundPublisher(repo Repository, publisher EventPublisher, logger logger.Logger, interval time.Duration, batchSize int) *BackgroundPublisher {
	return &BackgroundPublisher{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		interval:  interval,
		batchSize: batchSize,
		done:      make(chan struct{}),
	}
}

// Start runs the publisher loop until the context is canceled
func (bp *BackgroundPublisher) Start(ctx context.Context) {
	// Create cancellable context for graceful shutdown
	ctx, bp.cancel = context.WithCancel(ctx)
	defer close(bp.done)

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
		// Publish using service-specific publisher
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

// Close gracefully stops the background publisher
func (bp *BackgroundPublisher) Close() error {
	if bp.cancel != nil {
		bp.logger.Info("stopping background outbox publisher")
		bp.cancel()

		// Wait for goroutine to finish with timeout
		select {
		case <-bp.done:
			bp.logger.Info("background outbox publisher stopped gracefully")
		case <-time.After(5 * time.Second):
			bp.logger.Warn("background outbox publisher stop timeout")
		}
	}
	return nil
}
