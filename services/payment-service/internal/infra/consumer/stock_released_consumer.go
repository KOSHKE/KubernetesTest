package consumer

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"

	"google.golang.org/protobuf/proto"
)

// StockReleasedHandler defines interface for handling stock released events
type StockReleasedHandler interface {
	Handle(ctx context.Context, evt *events.StockReleased) error
}

// StockReleasedHandlerFunc allows functions to implement StockReleasedHandler interface
type StockReleasedHandlerFunc func(ctx context.Context, evt *events.StockReleased) error

func (f StockReleasedHandlerFunc) Handle(ctx context.Context, evt *events.StockReleased) error {
	return f(ctx, evt)
}

// StockReleasedConsumer is a type-safe consumer for StockReleased events
type StockReleasedConsumer struct {
	*kafkaclient.TypedConsumer[*events.StockReleased]
}

// NewStockReleasedConsumer creates a new StockReleasedConsumer
func NewStockReleasedConsumer(
	bootstrapServers, groupID, autoOffsetReset string,
	handler StockReleasedHandler,
) (*StockReleasedConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
	}

	unmarshal := func(data []byte) (*events.StockReleased, error) {
		var evt events.StockReleased
		if err := proto.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		return &evt, nil
	}

	typedConsumer, err := kafkaclient.NewTypedConsumer(config, handler, unmarshal)
	if err != nil {
		return nil, err
	}

	return &StockReleasedConsumer{TypedConsumer: typedConsumer}, nil
}

// WithLogger sets logger for the consumer
func (c *StockReleasedConsumer) WithLogger(l logger.Logger) *StockReleasedConsumer {
	c.TypedConsumer = c.TypedConsumer.WithLogger(l)
	return c
}
