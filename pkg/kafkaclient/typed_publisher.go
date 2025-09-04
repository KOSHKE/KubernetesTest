package kafkaclient

import (
	"context"

	"ecommerce-platform/pkg/logger"
)

// EventPublisher defines interface for publishing typed events
type EventPublisher[T any] interface {
	Publish(ctx context.Context, evt T) error
	WithLogger(l logger.Logger) EventPublisher[T]
	Close() error
}

// EventPublisherFunc allows functions to implement EventPublisher interface
type EventPublisherFunc[T any] func(ctx context.Context, evt T) error

func (f EventPublisherFunc[T]) Publish(ctx context.Context, evt T) error {
	return f(ctx, evt)
}

// TypedPublisher is a type-safe wrapper around base Publisher for specific event types
type TypedPublisher[T any] struct {
	base    Publisher
	topic   string
	marshal func(T) ([]byte, error)
}

// NewTypedPublisher creates a new typed publisher for specific event type
func NewTypedPublisher[T any](
	config PublisherConfig,
	topic string,
	marshal func(T) ([]byte, error),
) (*TypedPublisher[T], error) {
	base, err := NewKafkaPublisher(config)
	if err != nil {
		return nil, err
	}

	return &TypedPublisher[T]{
		base:    base,
		topic:   topic,
		marshal: marshal,
	}, nil
}

// WithLogger sets logger for typed publisher
func (tp *TypedPublisher[T]) WithLogger(l logger.Logger) EventPublisher[T] {
	tp.base = tp.base.WithLogger(l)
	return tp
}

// Close closes the typed publisher
func (tp *TypedPublisher[T]) Close() error {
	return tp.base.Close()
}

// Publish publishes a typed event to the configured topic
func (tp *TypedPublisher[T]) Publish(ctx context.Context, evt T) error {
	bytes, err := tp.marshal(evt)
	if err != nil {
		return err
	}
	return tp.base.Publish(ctx, tp.topic, bytes)
}
