package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/payment-service/internal/application/dto"
	"ecommerce-platform/services/payment-service/internal/domain/entities"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaymentProcessedPublisher is a type-safe publisher for PaymentProcessed events
type PaymentProcessedPublisher struct {
	publisher kafkaclient.EventPublisher[*events.PaymentProcessed]
}

// NewPaymentProcessedPublisher creates a new typed publisher for PaymentProcessed events
func NewPaymentProcessedPublisher(bootstrapServers, topic string, logger logger.Logger) (*PaymentProcessedPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: bootstrapServers,
		ClientID:         "payment-service",
	}

	// Create typed publisher with protobuf marshal function
	publisher, err := kafkaclient.NewTypedPublisher(
		config,
		topic,
		func(evt *events.PaymentProcessed) ([]byte, error) {
			return proto.Marshal(evt)
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &PaymentProcessedPublisher{publisher: publisher}, nil
}

// WithLogger sets logger for the publisher
func (p *PaymentProcessedPublisher) WithLogger(l logger.Logger) *PaymentProcessedPublisher {
	p.publisher = p.publisher.WithLogger(l)
	return p
}

// Close closes the publisher
func (p *PaymentProcessedPublisher) Close() error {
	return p.publisher.Close()
}

// PublishPaymentProcessed publishes a PaymentProcessed event from domain entity
func (p *PaymentProcessedPublisher) PublishPaymentProcessed(ctx context.Context, payment *entities.Payment, success bool, message string) error {
	// Convert domain entity to proto event
	event := &events.PaymentProcessed{
		OrderId:    payment.OrderID,
		PaymentId:  payment.ID,
		Success:    success,
		Message:    message,
		Amount:     payment.Amount.Amount,
		Currency:   payment.Amount.Currency.String(),
		OccurredAt: timestamppb.New(time.Now()),
	}

	return p.publisher.Publish(ctx, event)
}

// PublishFromOutbox publishes events directly from outbox
func (p *PaymentProcessedPublisher) PublishFromOutbox(ctx context.Context, event outbox.Event) error {
	switch event.Type {
	case "PaymentProcessed":
		return p.publishPaymentProcessedFromOutbox(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *PaymentProcessedPublisher) publishPaymentProcessedFromOutbox(ctx context.Context, event outbox.Event) error {
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
		Currency:   paymentEvent.Amount.Currency.String(),
		OccurredAt: timestamppb.New(time.Now()),
	}

	return p.publisher.Publish(ctx, protoEvent)
}
