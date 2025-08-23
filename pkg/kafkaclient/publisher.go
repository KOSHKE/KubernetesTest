package kafkaclient

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kubernetestest/ecommerce-platform/pkg/logger"
)

// Publisher defines interface for publishing messages
type Publisher interface {
	Publish(ctx context.Context, topic string, value []byte) error
	WithLogger(l logger.Logger) Publisher
	Close() error
}

// PublisherConfig holds publisher configuration
type PublisherConfig struct {
	BootstrapServers string
	ClientID         string
	Acks             string
	DeliveryTimeout  time.Duration
	FlushTimeout     time.Duration
}

// KafkaPublisher implements Publisher interface with optimized delivery handling
type KafkaPublisher struct {
	p        *kafka.Producer
	delivery chan kafka.Event
	log      logger.Logger
	config   PublisherConfig
}

// NewKafkaPublisher creates new publisher with optimized config
func NewKafkaPublisher(config PublisherConfig) (*KafkaPublisher, error) {
	if config.Acks == "" {
		config.Acks = "all"
	}
	if config.DeliveryTimeout == 0 {
		config.DeliveryTimeout = 30 * time.Second
	}
	if config.FlushTimeout == 0 {
		config.FlushTimeout = 5 * time.Second
	}

	conf := &kafka.ConfigMap{
		"bootstrap.servers":   config.BootstrapServers,
		"client.id":           config.ClientID,
		"acks":                config.Acks,
		"delivery.timeout.ms": int(config.DeliveryTimeout.Milliseconds()),
		"request.timeout.ms":  int(config.DeliveryTimeout.Milliseconds()),
		"linger.ms":           5,        // Batch messages for 5ms
		"batch.size":          16384,    // 16KB batch size
		"compression.type":    "snappy", // Enable compression
	}

	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}

	kp := &KafkaPublisher{
		p:        p,
		delivery: make(chan kafka.Event, 1000), // Increased buffer
		config:   config,
	}

	// Start delivery monitoring goroutine
	go kp.monitorDelivery()

	return kp, nil
}

// WithLogger sets logger for publisher
func (k *KafkaPublisher) WithLogger(l logger.Logger) Publisher {
	k.log = l
	return k
}

// Close gracefully shuts down publisher
func (k *KafkaPublisher) Close() error {
	// Flush remaining messages
	remaining := k.p.Flush(int(k.config.FlushTimeout.Milliseconds()))
	if remaining > 0 && k.log != nil {
		k.log.Warn("failed to flush messages during shutdown", "remaining", remaining)
	}

	// Close delivery channel
	close(k.delivery)

	// Close producer
	k.p.Close()

	return nil
}

// Publish sends message asynchronously with delivery confirmation
func (k *KafkaPublisher) Publish(ctx context.Context, topic string, value []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue with publish
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: value,
		Headers: []kafka.Header{
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	if err := k.p.Produce(msg, k.delivery); err != nil {
		if k.log != nil {
			k.log.Error("failed to produce message", "error", err, "topic", topic)
		}
		return err
	}

	if k.log != nil {
		k.log.Debug("message queued for delivery", "topic", topic, "size", len(value))
	}

	return nil
}

// monitorDelivery handles delivery confirmations and errors
func (k *KafkaPublisher) monitorDelivery() {
	for evt := range k.delivery {
		switch e := evt.(type) {
		case *kafka.Message:
			if e.TopicPartition.Error != nil {
				k.logDeliveryError(e)
			} else {
				k.logDeliverySuccess(e)
			}
		case *kafka.Error:
			if k.log != nil {
				k.log.Error("kafka producer error", "error", e.Error(), "code", e.Code())
			}
		}
	}
}

// logDeliveryError logs failed message delivery
func (k *KafkaPublisher) logDeliveryError(msg *kafka.Message) {
	if k.log == nil {
		return
	}

	topic := ""
	if msg.TopicPartition.Topic != nil {
		topic = *msg.TopicPartition.Topic
	}

	k.log.Error("message delivery failed",
		"topic", topic,
		"partition", msg.TopicPartition.Partition,
		"offset", msg.TopicPartition.Offset,
		"error", msg.TopicPartition.Error)
}

// logDeliverySuccess logs successful message delivery
func (k *KafkaPublisher) logDeliverySuccess(msg *kafka.Message) {
	if k.log == nil {
		return
	}

	topic := ""
	if msg.TopicPartition.Topic != nil {
		topic = *msg.TopicPartition.Topic
	}

	k.log.Debug("message delivered successfully",
		"topic", topic,
		"partition", msg.TopicPartition.Partition,
		"offset", msg.TopicPartition.Offset)
}
