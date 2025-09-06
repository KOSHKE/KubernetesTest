package outbox

import (
	"context"
	"time"

	"ecommerce-platform/pkg/logger"
)

// Event represents an event to be published via outbox pattern
type Event struct {
	ID           uint       `json:"id"`
	EventID      string     `json:"event_id"` // UUID for deduplication
	Type         string     `json:"type"`
	Payload      any        `json:"payload"`
	PayloadHash  string     `json:"payload_hash"` // Hash for fast deduplication
	Priority     int        `json:"priority"`     // Higher = more critical
	RetryCount   int        `json:"retry_count"`
	LastAttempt  time.Time  `json:"last_attempt"`
	CreatedAt    time.Time  `json:"created_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	FailedAt     *time.Time `json:"failed_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// Repository interface for outbox operations
type Repository interface {
	SaveEvent(ctx context.Context, event Event) error
	GetUnprocessedEvents(ctx context.Context, limit int) ([]Event, error)
	MarkAsProcessed(ctx context.Context, id uint) error
	MarkAsFailed(ctx context.Context, id uint, retryCount int, err error) error
	GetFailedEvents(ctx context.Context, limit int) ([]Event, error)
	MoveToDeadLetter(ctx context.Context, event Event, reason string) error
	GetDeadLetterEvents(ctx context.Context, limit int) ([]Event, error)
	RepublishFromDeadLetter(ctx context.Context, id uint) error
}

// Publisher interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

// OutboxRecord represents an outbox record in the database
type OutboxRecord struct {
	ID        uint
	Type      string
	Payload   string
	Processed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Service interface for outbox operations
type Service interface {
	SaveEvent(ctx context.Context, event Event) error
}

// OutboxService implements the Service interface
type OutboxService struct {
	repo   Repository
	logger logger.Logger
}

// NewService creates a new outbox service
func NewService(repo Repository, logger logger.Logger) Service {
	return &OutboxService{
		repo:   repo,
		logger: logger,
	}
}

// SaveEvent saves event to outbox table for later publishing
func (s *OutboxService) SaveEvent(ctx context.Context, event Event) error {
	err := s.repo.SaveEvent(ctx, event)
	if err != nil {
		s.logger.Error("failed to save event to outbox", "eventType", event.Type, "error", err)
		return err
	}

	return nil
}
