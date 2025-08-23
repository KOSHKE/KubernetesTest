package kafkaclient

import (
	"context"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kubernetestest/ecommerce-platform/pkg/logger"
)

// Consumer wraps a confluent-kafka consumer with optimized worker pool
type Consumer struct {
	c       *kafka.Consumer
	log     logger.Logger
	perMsgT time.Duration
	workers int
	buffer  int
}

// ConsumerConfig holds consumer configuration
type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	AutoOffsetReset  string
	Workers          int
	BufferSize       int
	HandleTimeout    time.Duration
}

// NewConsumer creates a new consumer with optimized config
func NewConsumer(config ConsumerConfig) (*Consumer, error) {
	if config.AutoOffsetReset == "" {
		config.AutoOffsetReset = "earliest"
	}
	if config.Workers <= 0 {
		config.Workers = 4
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 128
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       config.BootstrapServers,
		"group.id":                config.GroupID,
		"auto.offset.reset":       config.AutoOffsetReset,
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 1000,
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{
		c:       c,
		perMsgT: config.HandleTimeout,
		workers: config.Workers,
		buffer:  config.BufferSize,
	}, nil
}

// WithLogger sets logger for consumer
func (c *Consumer) WithLogger(l logger.Logger) *Consumer {
	c.log = l
	return c
}

// Close terminates consumer gracefully
func (c *Consumer) Close() error {
	return c.c.Close()
}

// RunValueLoop runs optimized worker-pool loop for value-only processing
func (c *Consumer) RunValueLoop(ctx context.Context, topics []string, handle func(context.Context, []byte) error) error {
	defer c.Close()

	if err := c.c.SubscribeTopics(topics, nil); err != nil {
		if c.log != nil {
			c.log.Error("failed to subscribe to topics", "error", err, "topics", topics)
		}
		return err
	}

	return c.runLoop(ctx, func(ctx context.Context, msg *kafka.Message) error {
		return handle(ctx, msg.Value)
	})
}

// RunMetaLoop runs optimized worker-pool loop with metadata
func (c *Consumer) RunMetaLoop(ctx context.Context, topics []string, handle func(context.Context, string, int32, kafka.Offset, []byte) error) error {
	defer c.Close()

	if err := c.c.SubscribeTopics(topics, nil); err != nil {
		if c.log != nil {
			c.log.Error("failed to subscribe to topics", "error", err, "topics", topics)
		}
		return err
	}

	return c.runLoop(ctx, func(ctx context.Context, msg *kafka.Message) error {
		topic := ""
		if msg.TopicPartition.Topic != nil {
			topic = *msg.TopicPartition.Topic
		}
		return handle(ctx, topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset, msg.Value)
	})
}

// runLoop is the unified worker pool implementation
func (c *Consumer) runLoop(ctx context.Context, handle func(context.Context, *kafka.Message) error) error {
	type workItem struct {
		msg *kafka.Message
	}

	workCh := make(chan workItem, c.buffer)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-workCh:
					if !ok {
						return
					}
					c.processMessage(ctx, item.msg, handle, workerID)
				}
			}
		}(i)
	}

	// Message reading loop
	for {
		select {
		case <-ctx.Done():
			close(workCh)
			wg.Wait()
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
			if msg == nil {
				continue
			}

			select {
			case workCh <- workItem{msg: msg}:
			case <-ctx.Done():
				close(workCh)
				wg.Wait()
				return nil
			default:
				if c.log != nil {
					c.log.Warn("dropping message due to full buffer", "topic", *msg.TopicPartition.Topic)
				}
			}
		}
	}
}

// processMessage handles individual message with timeout and error handling
func (c *Consumer) processMessage(ctx context.Context, msg *kafka.Message, handle func(context.Context, *kafka.Message) error, workerID int) {
	handleCtx := ctx
	var cancel func()

	if c.perMsgT > 0 {
		handleCtx, cancel = context.WithTimeout(ctx, c.perMsgT)
		defer cancel()
	}

	if err := handle(handleCtx, msg); err != nil {
		if c.log != nil {
			topic := ""
			if msg.TopicPartition.Topic != nil {
				topic = *msg.TopicPartition.Topic
			}
			c.log.Error("message processing failed",
				"error", err,
				"topic", topic,
				"partition", msg.TopicPartition.Partition,
				"offset", msg.TopicPartition.Offset,
				"worker_id", workerID)
		}
	} else if c.log != nil {
		topic := ""
		if msg.TopicPartition.Topic != nil {
			topic = *msg.TopicPartition.Topic
		}
		c.log.Debug("message processed successfully",
			"topic", topic,
			"partition", msg.TopicPartition.Partition,
			"offset", msg.TopicPartition.Offset,
			"worker_id", workerID)
	}
}
