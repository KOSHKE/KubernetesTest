package publisher

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	events "github.com/kubernetestest/ecommerce-platform/proto-go/events"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// SugaredLogger is a minimal interface compatible with zap.SugaredLogger
// and aligned with base publishers in this repo.
type SugaredLogger interface {
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

// PaymentProcessedPublisher publishes events.PaymentProcessed to a configured topic
type PaymentProcessedPublisher struct {
	p        *kafka.Producer
	delivery chan kafka.Event
	log      SugaredLogger
	topic    string
}

func NewPaymentProcessedPublisher(bootstrapServers, topic string) (*PaymentProcessedPublisher, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "payment-service",
		"acks":              "all",
	}
	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}
	pub := &PaymentProcessedPublisher{p: p, topic: topic, delivery: make(chan kafka.Event, 100)}
	// background drain with success/error logging
	go func() {
		for evt := range pub.delivery {
			if m, ok := evt.(*kafka.Message); ok {
				if m.TopicPartition.Error != nil {
					if pub.log != nil {
						t := ""
						if m.TopicPartition.Topic != nil {
							t = *m.TopicPartition.Topic
						}
						pub.log.Errorw("kafka delivery failed", "topic", t, "partition", m.TopicPartition.Partition, "offset", m.TopicPartition.Offset, "error", m.TopicPartition.Error)
					}
				} else {
					if pub.log != nil {
						t := ""
						if m.TopicPartition.Topic != nil {
							t = *m.TopicPartition.Topic
						}
						pub.log.Infow("kafka delivery succeeded", "topic", t, "partition", m.TopicPartition.Partition, "offset", m.TopicPartition.Offset)
					}
				}
			}
		}
	}()
	return pub, nil
}

// WithLogger attaches a structured logger (optional)
func (p *PaymentProcessedPublisher) WithLogger(l SugaredLogger) *PaymentProcessedPublisher {
	p.log = l
	return p
}

func (p *PaymentProcessedPublisher) Close() error {
	_ = p.p.Flush(5000)
	close(p.delivery)
	p.p.Close()
	return nil
}

func (p *PaymentProcessedPublisher) PublishPaymentProcessed(ctx context.Context, evt *events.PaymentProcessed) error {
	bytes, err := proto.Marshal(evt)
	if err != nil {
		return err
	}
	msg := &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny}, Value: bytes}
	if err := p.p.Produce(msg, p.delivery); err != nil {
		return err
	}
	// Non-blocking semantics with timeout/cancellation similar to base publishers
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	default:
		return nil
	}
}
