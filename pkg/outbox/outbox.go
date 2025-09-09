package outbox

import (
	"context"
	"encoding/json"
	"time"

	"ecommerce-platform/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Event represents a record in the outbox table
type Event struct {
	ID          uint
	AggregateID string
	Type        string
	Payload     interface{}
	RetryCount  int
	CreatedAt   time.Time
	ProcessedAt *time.Time
	FailedAt    *time.Time
	Error       string
}

// Repository describes work with the outbox table
type Repository interface {
	SaveEvent(ctx context.Context, e Event) error
	GetUnprocessedEvents(ctx context.Context, limit int) ([]Event, error)
	MarkAsProcessed(ctx context.Context, id uint) error
	MarkAsFailed(ctx context.Context, id uint, err string) error
}

// Service provides high-level outbox operations
type Service interface {
	SaveEvent(ctx context.Context, e Event) error
}

// OutboxService implements Service interface
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

// SaveEvent saves an event to the outbox
func (s *OutboxService) SaveEvent(ctx context.Context, e Event) error {
	if err := s.repo.SaveEvent(ctx, e); err != nil {
		s.logger.Error("failed to save event to outbox", "eventType", e.Type, "err", err)
		return err
	}
	return nil
}

// Publisher publishes events from outbox to Kafka via confluent-kafka-go
type Publisher struct {
	repo      Repository
	producer  *kafka.Producer
	topic     string
	logger    logger.Logger
	batchSize int
	interval  time.Duration
}

func NewPublisher(repo Repository, producer *kafka.Producer, topic string, logger logger.Logger, batchSize int, interval time.Duration) *Publisher {
	return &Publisher{
		repo:      repo,
		producer:  producer,
		topic:     topic,
		logger:    logger,
		batchSize: batchSize,
		interval:  interval,
	}
}

// Start runs the background worker
func (p *Publisher) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

// processBatch selects events from outbox and publishes them to Kafka
func (p *Publisher) processBatch(ctx context.Context) {
	events, err := p.repo.GetUnprocessedEvents(ctx, p.batchSize)
	if err != nil {
		p.logger.Error("failed to fetch events", "err", err)
		return
	}

	for _, e := range events {
		// Serialize payload to bytes
		payloadBytes, err := json.Marshal(e.Payload)
		if err != nil {
			p.logger.Error("failed to marshal payload", "eventID", e.ID, "err", err)
			_ = p.repo.MarkAsFailed(ctx, e.ID, err.Error())
			continue
		}

		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
			Key:            []byte(e.AggregateID),
			Value:          payloadBytes,
		}

		// Produce asynchronously, wait for delivery report via Events channel
		deliveryChan := make(chan kafka.Event, 1)
		err = p.producer.Produce(msg, deliveryChan)
		if err != nil {
			p.logger.Error("failed to produce", "eventID", e.ID, "err", err)
			_ = p.repo.MarkAsFailed(ctx, e.ID, err.Error())
			close(deliveryChan)
			continue
		}

		// Wait for confirmation from Kafka
		ev := <-deliveryChan
		m, ok := ev.(*kafka.Message)
		close(deliveryChan)

		if !ok {
			p.logger.Error("unexpected delivery report type", "eventID", e.ID)
			_ = p.repo.MarkAsFailed(ctx, e.ID, "unexpected delivery report")
			continue
		}

		if m.TopicPartition.Error != nil {
			p.logger.Error("delivery failed", "eventID", e.ID, "err", m.TopicPartition.Error)
			_ = p.repo.MarkAsFailed(ctx, e.ID, m.TopicPartition.Error.Error())
			continue
		}

		// Successfully delivered
		if err := p.repo.MarkAsProcessed(ctx, e.ID); err != nil {
			p.logger.Error("failed to mark processed", "eventID", e.ID, "err", err)
		}
	}
}
