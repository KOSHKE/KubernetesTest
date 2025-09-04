package publisher

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"

	"google.golang.org/protobuf/proto"

	"ecommerce-platform/proto-go/events"
)

// StockReservedPublisher is a type-safe publisher for StockReserved events
type StockReservedPublisher struct {
	publisher kafkaclient.EventPublisher[*events.StockReserved]
}

// NewStockReservedPublisher creates a new typed publisher for StockReserved events
func NewStockReservedPublisher(bootstrapServers, topic string) (*StockReservedPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: bootstrapServers,
		ClientID:         "inventory-service",
	}

	// Create typed publisher with protobuf marshal function
	publisher, err := kafkaclient.NewTypedPublisher(
		config,
		topic,
		func(evt *events.StockReserved) ([]byte, error) {
			return proto.Marshal(evt)
		},
	)
	if err != nil {
		return nil, err
	}

	return &StockReservedPublisher{publisher: publisher}, nil
}

// WithLogger sets logger for the publisher
func (p *StockReservedPublisher) WithLogger(l logger.Logger) *StockReservedPublisher {
	p.publisher = p.publisher.WithLogger(l)
	return p
}

// Close closes the publisher
func (p *StockReservedPublisher) Close() error {
	return p.publisher.Close()
}

// PublishStockReserved publishes a StockReserved event
func (p *StockReservedPublisher) PublishStockReserved(ctx context.Context, evt *events.StockReserved) error {
	return p.publisher.Publish(ctx, evt)
}
