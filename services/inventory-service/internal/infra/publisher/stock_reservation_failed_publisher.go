package publisher

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"

	"google.golang.org/protobuf/proto"

	"ecommerce-platform/proto-go/events"
)

// StockReservationFailedPublisher is a type-safe publisher for StockReservationFailed events
type StockReservationFailedPublisher struct {
	publisher kafkaclient.EventPublisher[*events.StockReservationFailed]
}

// NewStockReservationFailedPublisher creates a new typed publisher for StockReservationFailed events
func NewStockReservationFailedPublisher(bootstrapServers, topic string) (*StockReservationFailedPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: bootstrapServers,
		ClientID:         "inventory-service",
	}

	// Create typed publisher with protobuf marshal function
	publisher, err := kafkaclient.NewTypedPublisher(
		config,
		topic,
		func(evt *events.StockReservationFailed) ([]byte, error) {
			return proto.Marshal(evt)
		},
	)
	if err != nil {
		return nil, err
	}

	return &StockReservationFailedPublisher{publisher: publisher}, nil
}

// WithLogger sets logger for the publisher
func (p *StockReservationFailedPublisher) WithLogger(l logger.Logger) *StockReservationFailedPublisher {
	p.publisher = p.publisher.WithLogger(l)
	return p
}

// Close closes the publisher
func (p *StockReservationFailedPublisher) Close() error {
	return p.publisher.Close()
}

// PublishStockReservationFailed publishes a StockReservationFailed event
func (p *StockReservationFailedPublisher) PublishStockReservationFailed(ctx context.Context, evt *events.StockReservationFailed) error {
	return p.publisher.Publish(ctx, evt)
}
