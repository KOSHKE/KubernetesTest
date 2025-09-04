package kafkaclient

import (
	"context"

	"ecommerce-platform/pkg/logger"
)

// EventHandler defines interface for handling typed events
type EventHandler[T any] interface {
	Handle(ctx context.Context, evt T) error
}

// EventHandlerFunc allows functions to implement EventHandler interface
type EventHandlerFunc[T any] func(ctx context.Context, evt T) error

func (f EventHandlerFunc[T]) Handle(ctx context.Context, evt T) error {
	return f(ctx, evt)
}

// TypedConsumer is a type-safe wrapper around base Consumer for specific event types
type TypedConsumer[T any] struct {
	base      *Consumer
	handler   EventHandler[T]
	unmarshal func([]byte) (T, error)
}

// NewTypedConsumer creates a new typed consumer for specific event type
func NewTypedConsumer[T any](
	config ConsumerConfig,
	handler EventHandler[T],
	unmarshal func([]byte) (T, error),
) (*TypedConsumer[T], error) {
	base, err := NewConsumer(config)
	if err != nil {
		return nil, err
	}

	return &TypedConsumer[T]{
		base:      base,
		handler:   handler,
		unmarshal: unmarshal,
	}, nil
}

// WithLogger sets logger for typed consumer
func (tc *TypedConsumer[T]) WithLogger(l logger.Logger) *TypedConsumer[T] {
	tc.base.WithLogger(l)
	return tc
}

// Close closes the typed consumer
func (tc *TypedConsumer[T]) Close() error {
	return tc.base.Close()
}

// Run starts consuming messages with type-safe event handling
func (tc *TypedConsumer[T]) Run(ctx context.Context, topics []string) error {
	return tc.base.RunValueLoop(ctx, topics, func(hctx context.Context, value []byte) error {
		evt, err := tc.unmarshal(value)
		if err != nil {
			return err
		}
		return tc.handler.Handle(hctx, evt)
	})
}
