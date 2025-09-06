package outbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
)

// BackgroundPublisher handles publishing outbox events with workers, rate limiting, retries, and DLQ
type BackgroundPublisher struct {
	outboxRepo Repository
	publisher  Publisher
	logger     logger.Logger
	metrics    metrics.Metrics

	config      OutboxConfig
	rateLimiter *RateLimiter

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// OutboxConfig defines configuration for the publisher
type OutboxConfig struct {
	Workers      int
	Interval     time.Duration
	BatchSize    int
	EventTimeout time.Duration

	MaxRetries    int
	RetryDelay    time.Duration
	MaxRetryDelay time.Duration

	RateLimit  int
	BurstLimit int

	EnableDLQ     bool
	DLQRetryAfter time.Duration
}

// NewBackgroundPublisher creates a new instance
func NewBackgroundPublisher(
	repo Repository,
	pub Publisher,
	logger logger.Logger,
	metrics metrics.Metrics,
	cfg OutboxConfig,
) *BackgroundPublisher {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundPublisher{
		outboxRepo:  repo,
		publisher:   pub,
		logger:      logger,
		metrics:     metrics,
		config:      cfg,
		rateLimiter: NewRateLimiter(cfg.RateLimit, cfg.BurstLimit),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins processing with workers
func (bp *BackgroundPublisher) Start() {
	bp.logger.Info("starting outbox publisher",
		"workers", bp.config.Workers,
		"interval", bp.config.Interval,
		"batchSize", bp.config.BatchSize,
	)
	for i := 0; i < bp.config.Workers; i++ {
		bp.wg.Add(1)
		go bp.worker(i)
	}
}

// Stop gracefully stops all workers
func (bp *BackgroundPublisher) Stop() {
	bp.logger.Info("stopping outbox publisher")
	bp.cancel()
	bp.wg.Wait()
	bp.logger.Info("outbox publisher stopped")
}

// worker fetches and publishes events in a loop
func (bp *BackgroundPublisher) worker(id int) {
	defer bp.wg.Done()
	ticker := time.NewTicker(bp.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-bp.ctx.Done():
			bp.logger.Debug("worker stopped", "id", id)
			return
		case <-ticker.C:
			bp.processBatch()
		}
	}
}

// processBatch fetches and publishes a batch of events
func (bp *BackgroundPublisher) processBatch() {
	events, err := bp.outboxRepo.GetUnprocessedEvents(bp.ctx, bp.config.BatchSize)
	if err != nil {
		bp.logger.Error("failed to fetch events", "error", err)
		return
	}

	for _, e := range events {
		if !bp.rateLimiter.Allow() {
			bp.logger.Debug("rate limit reached, skipping event", "eventID", e.ID)
			continue
		}
		if err := bp.processEvent(e); err != nil {
			bp.logger.Error("failed to process event", "eventID", e.ID, "error", err)
		}
	}
}

// processEvent handles a single event with retry, timeout, and DLQ
func (bp *BackgroundPublisher) processEvent(event Event) error {
	attempt := 0
	for {
		attempt++
		ctx, cancel := context.WithTimeout(bp.ctx, bp.config.EventTimeout)
		err := bp.publisher.Publish(ctx, event)
		cancel()

		if err == nil {
			return bp.outboxRepo.MarkAsProcessed(bp.ctx, event.ID)
		}

		if attempt >= bp.config.MaxRetries {
			if bp.config.EnableDLQ {
				bp.logger.Warn("moving event to DLQ", "eventID", event.ID)
				return bp.outboxRepo.MoveToDLQ(bp.ctx, event.ID, err.Error())
			}
			return fmt.Errorf("max retries reached for event %d: %w", event.ID, err)
		}

		delay := bp.calculateRetryDelay(attempt)
		bp.logger.Debug("retrying event after delay", "eventID", event.ID, "delay", delay)
		time.Sleep(delay)
	}
}

// calculateRetryDelay uses exponential backoff
func (bp *BackgroundPublisher) calculateRetryDelay(attempt int) time.Duration {
	delay := float64(bp.config.RetryDelay) * math.Pow(2, float64(attempt-1))
	if delay > float64(bp.config.MaxRetryDelay) {
		delay = float64(bp.config.MaxRetryDelay)
	}
	return time.Duration(delay)
}

// generatePayloadHash generates unique hash for deduplication
func generatePayloadHash(payload any) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", payload)))
	return hex.EncodeToString(hash[:])
}

// --- Rate limiter ---
type RateLimiter struct {
	tokens chan struct{}
	rate   int
	burst  int
	ticker *time.Ticker
}

func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, burst),
		rate:   rate,
		burst:  burst,
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
	}

	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
}
