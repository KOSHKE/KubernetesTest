package kafkaclient

import (
	"context"

	"ecommerce-platform/pkg/logger"

	"github.com/segmentio/kafka-go"
)

// Consumer is a lightweight Kafka consumer with optional logging
type Consumer struct {
	reader *kafka.Reader
	log    logger.Logger
}

// ConsumerConfig holds minimal config
type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	AutoOffsetReset  string
}

// NewConsumer creates a simple consumer
func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	offset := kafka.FirstOffset
	if cfg.AutoOffsetReset == "latest" {
		offset = kafka.LastOffset
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.BootstrapServers},
		GroupID:     cfg.GroupID,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
		StartOffset: offset,
	})

	return &Consumer{reader: reader}, nil
}

// WithLogger sets logger for consumer
func (c *Consumer) WithLogger(l logger.Logger) *Consumer {
	c.log = l
	return c
}

// Close shuts down the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// Run consumes messages sequentially (no worker pool)
func (c *Consumer) Run(ctx context.Context, topics []string, handle func([]byte) error) error {
	// Subscribe to topics by creating a new reader for each topic
	// Segmentio kafka-go doesn't support multiple topics in one reader
	for _, topic := range topics {
		go func(topicName string) {
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:     []string{c.reader.Config().Brokers[0]},
				GroupID:     c.reader.Config().GroupID,
				Topic:       topicName,
				MinBytes:    1,
				MaxBytes:    10e6,
				StartOffset: c.reader.Config().StartOffset,
			})
			defer reader.Close()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					msg, err := reader.ReadMessage(ctx)
					if err != nil {
						if c.log != nil {
							c.log.Warn("kafka read error", "topic", topicName, "error", err)
						}
						continue
					}

					if err := handle(msg.Value); err != nil && c.log != nil {
						c.log.Error("message handling failed", "topic", topicName, "error", err)
					}
				}
			}
		}(topic)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}
