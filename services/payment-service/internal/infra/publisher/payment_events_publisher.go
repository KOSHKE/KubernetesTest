package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/payment-service/internal/application/dto"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaymentEventsPublisherImpl implements the PaymentEventsPublisher interface
type PaymentEventsPublisherImpl struct {
	publishers map[string]kafkaclient.EventPublisher[proto.Message]
	logger     logger.Logger
}

// NewPaymentEventsPublisher creates a new payment events publisher
func NewPaymentEventsPublisher(brokers []string, topics map[string]string, logger logger.Logger) (publisher.PaymentEventsPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: brokers[0],
		ClientID:         "payment-service",
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

	return &PaymentEventsPublisherImpl{
		publishers: publishers,
		logger:     logger,
	}, nil
}

// PublishPaymentProcessed publishes when payment is processed
func (p *PaymentEventsPublisherImpl) PublishPaymentProcessed(ctx context.Context, event *events.PaymentProcessed) error {
	publisher, exists := p.publishers["PaymentProcessed"]
	if !exists {
		return fmt.Errorf("publisher for PaymentProcessed not found")
	}
	return publisher.Publish(ctx, event)
}

// PublishFromOutbox publishes events directly from outbox
func (p *PaymentEventsPublisherImpl) PublishFromOutbox(ctx context.Context, event outbox.Event) error {
	switch event.Type {
	case "PaymentProcessed":
		return p.publishPaymentProcessedFromOutbox(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *PaymentEventsPublisherImpl) publishPaymentProcessedFromOutbox(ctx context.Context, event outbox.Event) error {
	// Parse payload to PaymentEventDTO
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var paymentEvent dto.PaymentEventDTO
	if err := json.Unmarshal(payloadBytes, &paymentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create protobuf event
	protoEvent := &events.PaymentProcessed{
		OrderId:    paymentEvent.OrderID,
		PaymentId:  paymentEvent.PaymentID,
		Success:    paymentEvent.Success,
		Message:    paymentEvent.Message,
		Amount:     paymentEvent.Amount.Amount,
		Currency:   paymentEvent.Amount.Currency.Code,
		OccurredAt: timestamppb.Now(),
	}

	return p.PublishPaymentProcessed(ctx, protoEvent)
}

// Close closes the publisher
func (p *PaymentEventsPublisherImpl) Close() error {
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
