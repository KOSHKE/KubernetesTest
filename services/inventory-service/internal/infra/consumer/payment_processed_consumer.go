package consumer

import (
	"context"

	"ecommerce-platform/pkg/kafkaclient"
	events "ecommerce-platform/proto-go/events"

	"google.golang.org/protobuf/proto"
)

// PaymentProcessedHandler defines interface for handling payment processed events
type PaymentProcessedHandler interface {
	Handle(ctx context.Context, evt *events.PaymentProcessed) error
}

// PaymentProcessedHandlerFunc allows functions to implement PaymentProcessedHandler interface
type PaymentProcessedHandlerFunc func(ctx context.Context, evt *events.PaymentProcessed) error

func (f PaymentProcessedHandlerFunc) Handle(ctx context.Context, evt *events.PaymentProcessed) error {
	return f(ctx, evt)
}

// PaymentProcessedConsumer is a type-safe consumer for PaymentProcessed events
type PaymentProcessedConsumer struct {
	*kafkaclient.TypedConsumer[*events.PaymentProcessed]
}

// NewPaymentProcessedConsumer creates a new PaymentProcessedConsumer
func NewPaymentProcessedConsumer(bootstrapServers, groupID, autoOffsetReset string, handler PaymentProcessedHandler) (*PaymentProcessedConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
	}

	unmarshal := func(data []byte) (*events.PaymentProcessed, error) {
		var evt events.PaymentProcessed
		if err := proto.Unmarshal(data, &evt); err != nil {
			return nil, err
		}
		return &evt, nil
	}

	typedConsumer, err := kafkaclient.NewTypedConsumer(config, handler, unmarshal)
	if err != nil {
		return nil, err
	}

	return &PaymentProcessedConsumer{TypedConsumer: typedConsumer}, nil
}
