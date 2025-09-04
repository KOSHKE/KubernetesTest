package consumer

import (
	"context"

	"ecommerce-platform/pkg/kafkaclient"
	events "ecommerce-platform/proto-go/events"
	"google.golang.org/protobuf/proto"
)

// StockReservedHandler defines interface for handling stock reserved events
type StockReservedHandler interface {
	Handle(ctx context.Context, evt *events.StockReserved) error
}

// StockReservedHandlerFunc allows functions to implement StockReservedHandler interface
type StockReservedHandlerFunc func(ctx context.Context, evt *events.StockReserved) error

func (f StockReservedHandlerFunc) Handle(ctx context.Context, evt *events.StockReserved) error {
	return f(ctx, evt)
}

// StockReservedConsumer is a type-safe consumer for StockReserved events
type StockReservedConsumer struct {
	*kafkaclient.TypedConsumer[*events.StockReserved]
}

// NewStockReservedConsumer creates a new StockReservedConsumer
func NewStockReservedConsumer(bootstrapServers, groupID, autoOffsetReset string, handler StockReservedHandler) (*StockReservedConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
	}

	unmarshal := func(data []byte) (*events.StockReserved, error) {
		var evt events.StockReserved
		if err := proto.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		return &evt, nil
	}

	typedConsumer, err := kafkaclient.NewTypedConsumer(config, handler, unmarshal)
	if err != nil {
		return nil, err
	}

	return &StockReservedConsumer{TypedConsumer: typedConsumer}, nil
}
