package kafkaclient

import (
	"context"

	"ecommerce-platform/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Publisher defines minimal interface for sending messages
type Publisher interface {
	Publish(ctx context.Context, topic string, value []byte) error
	Close() error
}

// KafkaPublisher is a simplified Kafka publisher
type KafkaPublisher struct {
	producer *kafka.Producer
	log      logger.Logger
}

// NewKafkaPublisher creates a minimal KafkaPublisher
func NewKafkaPublisher(bootstrapServers string, clientID string, log logger.Logger) (*KafkaPublisher, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         clientID,
		"acks":              "all",
		"linger.ms":         5,
		"compression.type":  "snappy",
	}

	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}

	return &KafkaPublisher{
		producer: p,
		log:      log,
	}, nil
}

// Publish sends a message asynchronously
func (k *KafkaPublisher) Publish(ctx context.Context, topic string, value []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value,
	}

	if err := k.producer.Produce(msg, nil); err != nil {
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

// WithLogger sets logger for publisher
func (k *KafkaPublisher) WithLogger(l logger.Logger) *KafkaPublisher {
	k.log = l
	return k
}

// Close flushes and closes the producer
func (k *KafkaPublisher) Close() error {
	k.producer.Flush(5000) // wait up to 5s for delivery
	k.producer.Close()
	return nil
}
