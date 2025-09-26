package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	dto "ecommerce-platform/pkg/common/dto/inventory-service"
	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/publisher"

	"google.golang.org/protobuf/proto"
)

// StockEventsPublisherImpl implements the StockEventsPublisher interface
type StockEventsPublisherImpl struct {
	publishers map[string]kafkaclient.EventPublisher[proto.Message]
	logger     logger.Logger
}

// NewStockEventsPublisher creates a new stock events publisher
func NewStockEventsPublisher(brokers []string, topics map[string]string, logger logger.Logger) (publisher.StockEventsPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: brokers[0],
		ClientID:         "inventory-service",
	}

	publishers := make(map[string]kafkaclient.EventPublisher[proto.Message])

	// Create publishers for each event type
	for eventType, topic := range topics {
		pub, err := kafkaclient.NewTypedPublisher(
			config,
			topic,
			func(msg proto.Message) ([]byte, error) {
				return proto.Marshal(msg)
			},
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create publisher for %s: %w", eventType, err)
		}
		publishers[eventType] = pub
	}

	return &StockEventsPublisherImpl{
		publishers: publishers,
		logger:     logger,
	}, nil
}

// PublishStockReserved publishes when stock is successfully reserved
func (p *StockEventsPublisherImpl) PublishStockReserved(ctx context.Context, event *events.StockReserved) error {
	publisher, exists := p.publishers["StockReserved"]
	if !exists {
		return fmt.Errorf("publisher for StockReserved not found")
	}
	return publisher.Publish(ctx, event)
}

// PublishStockReleased publishes when reserved stock is released
func (p *StockEventsPublisherImpl) PublishStockReleased(ctx context.Context, event *events.StockReleased) error {
	publisher, exists := p.publishers["StockReleased"]
	if !exists {
		return fmt.Errorf("publisher for StockReleased not found")
	}
	return publisher.Publish(ctx, event)
}

// PublishStockCommitted publishes when reserved stock is committed
func (p *StockEventsPublisherImpl) PublishStockCommitted(ctx context.Context, event *events.StockCommitted) error {
	publisher, exists := p.publishers["StockCommitted"]
	if !exists {
		return fmt.Errorf("publisher for StockCommitted not found")
	}
	return publisher.Publish(ctx, event)
}

// PublishFromOutbox publishes events directly from outbox
func (p *StockEventsPublisherImpl) PublishFromOutbox(ctx context.Context, event outbox.Event) error {
	switch event.Type {
	case "StockReserved":
		return p.publishStockReservedFromOutbox(ctx, event)
	case "StockReleased":
		return p.publishStockReleasedFromOutbox(ctx, event)
	case "StockCommitted":
		return p.publishStockCommittedFromOutbox(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *StockEventsPublisherImpl) publishStockReservedFromOutbox(ctx context.Context, event outbox.Event) error {
	// Parse payload to StockEventDTO
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var stockEvent dto.StockEventDTO
	if err := json.Unmarshal(payloadBytes, &stockEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create protobuf event
	protoEvent := &events.StockReserved{
		OrderId: event.AggregateID,
	}

	return p.PublishStockReserved(ctx, protoEvent)
}

func (p *StockEventsPublisherImpl) publishStockReleasedFromOutbox(ctx context.Context, event outbox.Event) error {
	// Parse payload to StockEventDTO
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var stockEvent dto.StockEventDTO
	if err := json.Unmarshal(payloadBytes, &stockEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create protobuf event
	protoEvent := &events.StockReleased{
		OrderId: event.AggregateID,
	}

	return p.PublishStockReleased(ctx, protoEvent)
}

func (p *StockEventsPublisherImpl) publishStockCommittedFromOutbox(ctx context.Context, event outbox.Event) error {
	// Parse payload to StockEventDTO
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var stockEvent dto.StockEventDTO
	if err := json.Unmarshal(payloadBytes, &stockEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create protobuf event
	protoEvent := &events.StockCommitted{
		OrderId: event.AggregateID,
	}

	return p.PublishStockCommitted(ctx, protoEvent)
}

// Close closes the publisher
func (p *StockEventsPublisherImpl) Close() error {
	var firstErr error
	for eventType, publisher := range p.publishers {
		if err := publisher.Close(); err != nil {
			p.logger.Error("failed to close publisher", "eventType", eventType, "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
