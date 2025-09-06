package consumer

import (
	"context"

	kafkaclient "ecommerce-platform/pkg/kafkaclient"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"

	"google.golang.org/protobuf/proto"
)

// StockCommittedHandler defines interface for handling stock committed events
type StockCommittedHandler interface {
	Handle(ctx context.Context, evt *events.StockCommitted) error
}

// StockCommittedHandlerFunc allows functions to implement StockCommittedHandler interface
type StockCommittedHandlerFunc func(ctx context.Context, evt *events.StockCommitted) error

func (f StockCommittedHandlerFunc) Handle(ctx context.Context, evt *events.StockCommitted) error {
	return f(ctx, evt)
}

// StockCommittedConsumer is a type-safe consumer for StockCommitted events
type StockCommittedConsumer struct {
	*kafkaclient.TypedConsumer[*events.StockCommitted]
}

// NewStockCommittedConsumer creates a new StockCommittedConsumer
func NewStockCommittedConsumer(
	bootstrapServers, groupID, autoOffsetReset string,
	handler StockCommittedHandler,
) (*StockCommittedConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
	}

	unmarshal := func(data []byte) (*events.StockCommitted, error) {
		var evt events.StockCommitted
		if err := proto.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		return &evt, nil
	}

	typedConsumer, err := kafkaclient.NewTypedConsumer(config, handler, unmarshal)
	if err != nil {
		return nil, err
	}

	return &StockCommittedConsumer{TypedConsumer: typedConsumer}, nil
}

// WithLogger sets logger for the consumer
func (c *StockCommittedConsumer) WithLogger(l logger.Logger) *StockCommittedConsumer {
	c.TypedConsumer = c.TypedConsumer.WithLogger(l)
	return c
}
