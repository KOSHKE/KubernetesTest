package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	dto "ecommerce-platform/pkg/common/dto/order-service"
	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/order-service/internal/domain/ports/publisher"

	"google.golang.org/protobuf/proto"
)

// OrderEventsPublisherImpl implements the OrderEventsPublisher interface
type OrderEventsPublisherImpl struct {
	publishers map[string]kafkaclient.EventPublisher[proto.Message]
	logger     logger.Logger
}

// NewOrderEventsPublisher creates a new order events publisher
func NewOrderEventsPublisher(brokers []string, topics map[string]string, logger logger.Logger) (publisher.OrderEventsPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: brokers[0],
		ClientID:         "order-service",
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

	return &OrderEventsPublisherImpl{
		publishers: publishers,
		logger:     logger,
	}, nil
}

// PublishOrderCreated publishes when an order is created
func (p *OrderEventsPublisherImpl) PublishOrderCreated(ctx context.Context, event *events.OrderCreated) error {
	publisher, exists := p.publishers["OrderCreated"]
	if !exists {
		return fmt.Errorf("publisher for OrderCreated not found")
	}
	return publisher.Publish(ctx, event)
}

// PublishOrderCancelled publishes when an order is cancelled
func (p *OrderEventsPublisherImpl) PublishOrderCancelled(ctx context.Context, event *events.OrderCancelled) error {
	publisher, exists := p.publishers["OrderCancelled"]
	if !exists {
		return fmt.Errorf("publisher for OrderCancelled not found")
	}
	return publisher.Publish(ctx, event)
}

// PublishFromOutbox publishes events directly from outbox
func (p *OrderEventsPublisherImpl) PublishFromOutbox(ctx context.Context, event outbox.Event) error {
	switch event.Type {
	case "OrderCreated":
		return p.publishOrderCreatedFromOutbox(ctx, event)
	case "OrderCancelled":
		return p.publishOrderCancelledFromOutbox(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *OrderEventsPublisherImpl) publishOrderCreatedFromOutbox(ctx context.Context, event outbox.Event) error {
	// Parse payload to OrderEventDTO
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var orderEvent dto.OrderEventDTO
	if err := json.Unmarshal(payloadBytes, &orderEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create protobuf event
	protoEvent := &events.OrderCreated{
		OrderId:     event.AggregateID,
		UserId:      orderEvent.UserID,
		Items:       orderEvent.Items,
		TotalAmount: orderEvent.TotalAmount,
		Currency:    orderEvent.Currency,
	}

	return p.PublishOrderCreated(ctx, protoEvent)
}

func (p *OrderEventsPublisherImpl) publishOrderCancelledFromOutbox(ctx context.Context, event outbox.Event) error {
	// Parse payload to OrderEventDTO
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var orderEvent dto.OrderEventDTO
	if err := json.Unmarshal(payloadBytes, &orderEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create protobuf event
	protoEvent := &events.OrderCancelled{
		OrderId: event.AggregateID,
		UserId:  orderEvent.UserID,
		Items:   orderEvent.Items,
		Reason:  orderEvent.Reason,
	}

	return p.PublishOrderCancelled(ctx, protoEvent)
}

// Close closes the publisher
func (p *OrderEventsPublisherImpl) Close() error {
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
