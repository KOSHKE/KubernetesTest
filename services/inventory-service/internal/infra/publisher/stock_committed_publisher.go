package publisher

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"

	"google.golang.org/protobuf/proto"

	"ecommerce-platform/proto-go/events"
)

// StockCommittedPublisher is a type-safe publisher for StockCommitted events
type StockCommittedPublisher struct {
	publisher kafkaclient.EventPublisher[*events.StockCommitted]
}

// NewStockCommittedPublisher creates a new typed publisher for StockCommitted events
func NewStockCommittedPublisher(bootstrapServers, topic string) (*StockCommittedPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: bootstrapServers,
		ClientID:         "inventory-service",
	}

	// Create typed publisher with protobuf marshal function
	publisher, err := kafkaclient.NewTypedPublisher(
		config,
		topic,
		func(evt *events.StockCommitted) ([]byte, error) {
			return proto.Marshal(evt)
		},
	)
	if err != nil {
		return nil, err
	}

	return &StockCommittedPublisher{publisher: publisher}, nil
}

// WithLogger sets logger for the publisher
func (p *StockCommittedPublisher) WithLogger(l logger.Logger) *StockCommittedPublisher {
	p.publisher = p.publisher.WithLogger(l)
	return p
}

// Close closes the publisher
func (p *StockCommittedPublisher) Close() error {
	return p.publisher.Close()
}

// PublishStockCommitted publishes a StockCommitted event
func (p *StockCommittedPublisher) PublishStockCommitted(ctx context.Context, evt *events.StockCommitted) error {
	return p.publisher.Publish(ctx, evt)
}
