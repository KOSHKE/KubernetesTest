package kafkaclient

import (
	"context"
	"time"

	"ecommerce-platform/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Consumer is a lightweight Kafka consumer with optional logging
type Consumer struct {
	c   *kafka.Consumer
	log logger.Logger
}

// ConsumerConfig holds minimal config
type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	AutoOffsetReset  string
}

// NewConsumer creates a simple consumer
func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	if cfg.AutoOffsetReset == "" {
		cfg.AutoOffsetReset = "earliest"
	}

	kc, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.BootstrapServers,
		"group.id":          cfg.GroupID,
		"auto.offset.reset": cfg.AutoOffsetReset,
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{c: kc}, nil
}

// WithLogger sets logger for consumer
func (c *Consumer) WithLogger(l logger.Logger) *Consumer {
	c.log = l
	return c
}

// Close shuts down the consumer
func (c *Consumer) Close() error {
	return c.c.Close()
}

// Run consumes messages sequentially (no worker pool)
func (c *Consumer) Run(ctx context.Context, topics []string, handle func([]byte) error) error {
	if err := c.c.SubscribeTopics(topics, nil); err != nil {
		if c.log != nil {
			c.log.Error("failed to subscribe topics", "error", err)
		}
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kerr, ok := err.(kafka.Error); !ok || kerr.Code() != kafka.ErrTimedOut {
					if c.log != nil {
						c.log.Warn("kafka read error", "error", err)
					}
				}
				continue
			}

			if err := handle(msg.Value); err != nil && c.log != nil {
				c.log.Error("message handling failed", "error", err)
			}
		}
	}
}
