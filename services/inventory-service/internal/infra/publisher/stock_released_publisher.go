package publisher

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"

	"google.golang.org/protobuf/proto"

	"ecommerce-platform/proto-go/events"
)

// StockReleasedPublisher is a type-safe publisher for StockReleased events
type StockReleasedPublisher struct {
	publisher kafkaclient.EventPublisher[*events.StockReleased]
}

// NewStockReleasedPublisher creates a new typed publisher for StockReleased events
func NewStockReleasedPublisher(bootstrapServers, topic string) (*StockReleasedPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: bootstrapServers,
		ClientID:         "inventory-service",
	}

	// Create typed publisher with protobuf marshal function
	publisher, err := kafkaclient.NewTypedPublisher(
		config,
		topic,
		func(evt *events.StockReleased) ([]byte, error) {
			return proto.Marshal(evt)
		},
	)
	if err != nil {
		return nil, err
	}

	return &StockReleasedPublisher{publisher: publisher}, nil
}

// WithLogger sets logger for the publisher
func (p *StockReleasedPublisher) WithLogger(l logger.Logger) *StockReleasedPublisher {
	p.publisher = p.publisher.WithLogger(l)
	return p
}

// Close closes the publisher
func (p *StockReleasedPublisher) Close() error {
	return p.publisher.Close()
}

// PublishStockReleased publishes a StockReleased event
func (p *StockReleasedPublisher) PublishStockReleased(ctx context.Context, evt *events.StockReleased) error {
	return p.publisher.Publish(ctx, evt)
}
