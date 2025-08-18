package kafka

import (
	"context"
	"time"

	events "proto-go/events"

	"google.golang.org/protobuf/proto"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type PaymentProcessedHandler interface {
	Handle(ctx context.Context, evt *events.PaymentProcessed) error
}

type PaymentProcessedHandlerFunc func(ctx context.Context, evt *events.PaymentProcessed) error

func (f PaymentProcessedHandlerFunc) Handle(ctx context.Context, evt *events.PaymentProcessed) error {
	return f(ctx, evt)
}

type PaymentConsumer struct {
	c *ckafka.Consumer
	h PaymentProcessedHandler
}

func NewPaymentConsumer(bootstrapServers, groupID string, handler PaymentProcessedHandler) (*PaymentConsumer, error) {
	c, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	return &PaymentConsumer{c: c, h: handler}, nil
}

func (c *PaymentConsumer) Close() { _ = c.c.Close() }

func (c *PaymentConsumer) Run(ctx context.Context, topic string) error {
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
			var evt events.PaymentProcessed
			if unmarshalErr := proto.Unmarshal(msg.Value, &evt); unmarshalErr != nil {
				continue
			}
			_ = c.h.Handle(ctx, &evt)
		}
	}
}
