package kafka

import (
	"context"
	"time"

	events "proto-go/events"

	"google.golang.org/protobuf/proto"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type OrderCreatedHandler interface {
	Handle(ctx context.Context, evt *events.OrderCreated) error
}

// OrderCreatedHandlerFunc is a functional adapter to allow using ordinary functions as handlers.
type OrderCreatedHandlerFunc func(ctx context.Context, evt *events.OrderCreated) error

func (f OrderCreatedHandlerFunc) Handle(ctx context.Context, evt *events.OrderCreated) error {
	return f(ctx, evt)
}

type Consumer struct {
	c *kafka.Consumer
	h OrderCreatedHandler
}

func NewConsumer(bootstrapServers, groupID string, handler OrderCreatedHandler) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	return &Consumer{c: c, h: handler}, nil
}

func (c *Consumer) Close() { _ = c.c.Close() }

func (c *Consumer) Run(ctx context.Context, topic string) error {
	if err := c.c.SubscribeTopics([]string{topic}, nil); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.c.ReadMessage(100 * time.Millisecond)
			if err != nil || msg == nil {
				continue
			}
			var evt events.OrderCreated
			if unmarshalErr := proto.Unmarshal(msg.Value, &evt); unmarshalErr != nil {
				continue
			}
			_ = c.h.Handle(ctx, &evt)
		}
	}
}
