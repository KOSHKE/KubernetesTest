package publisher

import (
	"context"
	"time"

	"ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/payment-service/internal/domain/entities"

	"google.golang.org/protobuf/proto"
)

// PaymentProcessedPublisher is a type-safe publisher for PaymentProcessed events
type PaymentProcessedPublisher struct {
	publisher kafkaclient.EventPublisher[*events.PaymentProcessed]
}

// NewPaymentProcessedPublisher creates a new typed publisher for PaymentProcessed events
func NewPaymentProcessedPublisher(bootstrapServers, topic string) (*PaymentProcessedPublisher, error) {
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
		OccurredAt: time.Now().Format(time.RFC3339),
	}

	return p.publisher.Publish(ctx, event)
}
