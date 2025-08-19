package kafka

import (
	"context"
	"log"
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

type Consumer struct {
	c *ckafka.Consumer
	h PaymentProcessedHandler
}

func NewConsumer(bootstrapServers, groupID, autoOffsetReset string, handler PaymentProcessedHandler) (*Consumer, error) {
	if autoOffsetReset == "" {
		autoOffsetReset = "earliest"
	}
	c, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupID,
		"auto.offset.reset": autoOffsetReset,
	})
	if err != nil {
		return nil, err
	}
	return &Consumer{c: c, h: handler}, nil
}

func (c *Consumer) Close() { _ = c.c.Close() }

func (c *Consumer) Run(ctx context.Context, topics []string) error {
	defer c.Close()
	if err := c.c.SubscribeTopics(topics, nil); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kerr, ok := err.(ckafka.Error); ok {
					if kerr.Code() != ckafka.ErrTimedOut {
						log.Printf("kafka read error: %v", err)
					}
				} else {
					log.Printf("kafka read error: %v", err)
				}
				continue
			}
			if msg == nil {
				continue
			}
			var evt events.PaymentProcessed
			if unmarshalErr := proto.Unmarshal(msg.Value, &evt); unmarshalErr != nil {
				log.Printf("proto unmarshal error (topic=%s partition=%d offset=%d): %v", *msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset, unmarshalErr)
				continue
			}
			handleCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err := c.h.Handle(handleCtx, &evt); err != nil {
				log.Printf("handler error (topic=%s partition=%d offset=%d): %v", *msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset, err)
			}
			cancel()
		}
	}
}
