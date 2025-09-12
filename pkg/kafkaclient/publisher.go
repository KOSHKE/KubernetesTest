package kafkaclient

import (
	"context"
	"sync"

	"ecommerce-platform/pkg/logger"

	"github.com/segmentio/kafka-go"
)

// Publisher defines minimal interface for sending messages
type Publisher interface {
	Publish(ctx context.Context, topic string, value []byte) error
	Close() error
}

// KafkaPublisher is a simplified Kafka publisher with thread-safe writers
type KafkaPublisher struct {
	writers          map[string]*kafka.Writer
	bootstrapServers string
	log              logger.Logger
	mu               sync.Mutex // protects writers map
}

// NewKafkaPublisher creates a minimal KafkaPublisher
func NewKafkaPublisher(bootstrapServers string, clientID string, log logger.Logger) *KafkaPublisher {
	return &KafkaPublisher{
		writers:          make(map[string]*kafka.Writer),
		bootstrapServers: bootstrapServers,
		log:              log,
	}
}

// WithLogger sets a custom logger
func (k *KafkaPublisher) WithLogger(l logger.Logger) *KafkaPublisher {
	k.log = l
	return k
}

// Publish sends a message to the specified topic in a thread-safe way
func (k *KafkaPublisher) Publish(ctx context.Context, topic string, value []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Thread-safe access to the writers map
	k.mu.Lock()
	writer, exists := k.writers[topic]
	if !exists {
		// Create a new writer for this topic if it doesn't exist
		writer = &kafka.Writer{
			Addr:         kafka.TCP(k.bootstrapServers),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    1,                // immediate send, low latency
			RequiredAcks: kafka.RequireAll, // wait for all replicas
			Compression:  kafka.Snappy,     // compress messages
		}
		k.writers[topic] = writer
	}
	k.mu.Unlock()

	msg := kafka.Message{
		Value: value,
	}

	// Synchronous send
	if err := writer.WriteMessages(ctx, msg); err != nil {
		if k.log != nil {
			k.log.Error("failed to produce message", "topic", topic, "error", err)
		}
		return err
	}

	if k.log != nil {
		k.log.Debug("message queued for delivery", "topic", topic)
	}

	return nil
}

// Close flushes and closes all writers safely
func (k *KafkaPublisher) Close() error {
	var firstErr error
	k.mu.Lock()
	defer k.mu.Unlock()

	for topic, writer := range k.writers {
		if err := writer.Close(); err != nil {
			if k.log != nil {
				k.log.Error("failed to close writer", "topic", topic, "error", err)
			}
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
