package publisher

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"google.golang.org/protobuf/proto"

	"ecommerce-platform/proto-go/events"
)

// OrderCreatedPublisher is a type-safe publisher for OrderCreated events
type OrderCreatedPublisher struct {
	publisher kafkaclient.EventPublisher[*events.OrderCreated]
}

// NewOrderCreatedPublisher creates a new typed publisher for OrderCreated events
func NewOrderCreatedPublisher(bootstrapServers, topic string) (*OrderCreatedPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: bootstrapServers,
		ClientID:         "order-service",
	}

	// Create typed publisher with protobuf marshal function
	publisher, err := kafkaclient.NewTypedPublisher(
		config,
		topic,
		func(evt *events.OrderCreated) ([]byte, error) {
			return proto.Marshal(evt)
		},
	)
	if err != nil {
		return nil, err
	}

	return &OrderCreatedPublisher{publisher: publisher}, nil
}

// WithLogger sets logger for the publisher
func (p *OrderCreatedPublisher) WithLogger(l logger.Logger) *OrderCreatedPublisher {
	p.publisher = p.publisher.WithLogger(l)
	return p
}

// Close closes the publisher
func (p *OrderCreatedPublisher) Close() error {
	return p.publisher.Close()
}

// PublishOrderCreated publishes an OrderCreated event
func (p *OrderCreatedPublisher) PublishOrderCreated(ctx context.Context, evt *events.OrderCreated) error {
	return p.publisher.Publish(ctx, evt)
}
