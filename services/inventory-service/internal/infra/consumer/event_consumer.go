package consumer

import (
	"context"
	"reflect"

	"ecommerce-platform/pkg/kafkaclient"

	"google.golang.org/protobuf/proto"
)

// EventHandler defines interface for handling events
type EventHandler[T any] interface {
	Handle(ctx context.Context, evt T) error
}

// EventHandlerFunc allows functions to implement EventHandler interface
type EventHandlerFunc[T any] func(ctx context.Context, evt T) error

func (f EventHandlerFunc[T]) Handle(ctx context.Context, evt T) error {
	return f(ctx, evt)
}

// EventConsumer is a generic type-safe consumer for any protobuf event
type EventConsumer[T proto.Message] struct {
	*kafkaclient.TypedConsumer[T]
}

// NewEventConsumer creates a new generic EventConsumer
func NewEventConsumer[T proto.Message](
	bootstrapServers, groupID, autoOffsetReset string,
	handler EventHandler[T],
) (*EventConsumer[T], error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
	}

	unmarshal := func(data []byte) (T, error) {
		var zero T
		// T is proto.Message (обычно указатель). Создаём новый экземпляр через reflect.
		msg := reflect.New(reflect.TypeOf(zero).Elem()).Interface().(T)
		if err := proto.Unmarshal(data, msg); err != nil {
			return zero, err
		}
		return msg, nil
	}

	typedConsumer, err := kafkaclient.NewTypedConsumer(config, handler, unmarshal)
	if err != nil {
		return nil, err
	}

	return &EventConsumer[T]{TypedConsumer: typedConsumer}, nil
}
