package consumer

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"

	"google.golang.org/protobuf/proto"
)

// StockReservationFailedHandler defines interface for handling stock reservation failed events
type StockReservationFailedHandler interface {
	Handle(ctx context.Context, evt *events.StockReservationFailed) error
}

// StockReservationFailedHandlerFunc allows functions to implement StockReservationFailedHandler interface
type StockReservationFailedHandlerFunc func(ctx context.Context, evt *events.StockReservationFailed) error

func (f StockReservationFailedHandlerFunc) Handle(ctx context.Context, evt *events.StockReservationFailed) error {
	return f(ctx, evt)
}

// StockReservationFailedConsumer is a type-safe consumer for StockReservationFailed events
type StockReservationFailedConsumer struct {
	*kafkaclient.TypedConsumer[*events.StockReservationFailed]
}

// NewStockReservationFailedConsumer creates a new StockReservationFailedConsumer
func NewStockReservationFailedConsumer(
	bootstrapServers, groupID, autoOffsetReset string,
	handler StockReservationFailedHandler,
) (*StockReservationFailedConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
	}

	unmarshal := func(data []byte) (*events.StockReservationFailed, error) {
		var evt events.StockReservationFailed
		if err := proto.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		return &evt, nil
	}

	typedConsumer, err := kafkaclient.NewTypedConsumer(config, handler, unmarshal)
	if err != nil {
		return nil, err
	}

	return &StockReservationFailedConsumer{TypedConsumer: typedConsumer}, nil
}

// WithLogger sets logger for the consumer
func (c *StockReservationFailedConsumer) WithLogger(l logger.Logger) *StockReservationFailedConsumer {
	c.TypedConsumer = c.TypedConsumer.WithLogger(l)
	return c
}
