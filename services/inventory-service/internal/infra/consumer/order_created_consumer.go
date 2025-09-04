package consumer

import (
	"context"

	"ecommerce-platform/pkg/kafkaclient"
	events "ecommerce-platform/proto-go/events"

	"google.golang.org/protobuf/proto"
)

// OrderCreatedHandler defines interface for handling order created events
type OrderCreatedHandler interface {
	Handle(ctx context.Context, evt *events.OrderCreated) error
}

// OrderCreatedHandlerFunc allows functions to implement OrderCreatedHandler interface
type OrderCreatedHandlerFunc func(ctx context.Context, evt *events.OrderCreated) error

func (f OrderCreatedHandlerFunc) Handle(ctx context.Context, evt *events.OrderCreated) error {
	return f(ctx, evt)
}

// OrderCreatedConsumer is a type-safe consumer for OrderCreated events
type OrderCreatedConsumer struct {
	*kafkaclient.TypedConsumer[*events.OrderCreated]
}

// NewOrderCreatedConsumer creates a new OrderCreatedConsumer
func NewOrderCreatedConsumer(bootstrapServers, groupID, autoOffsetReset string, handler OrderCreatedHandler) (*OrderCreatedConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
	}

	unmarshal := func(data []byte) (*events.OrderCreated, error) {
		var evt events.OrderCreated
		if err := proto.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		return &evt, nil
	}

	typedConsumer, err := kafkaclient.NewTypedConsumer(config, handler, unmarshal)
	if err != nil {
		return nil, err
	}

	return &OrderCreatedConsumer{TypedConsumer: typedConsumer}, nil
}
