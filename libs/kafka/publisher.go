package kafka

import (
	"context"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Publisher defines a generic interface for publishing messages
type Publisher interface {
	Publish(ctx context.Context, topic string, value []byte) error
	WithLogger(l SugaredLogger) Publisher
	Close() error
}

type KafkaPublisher struct {
	p        *ckafka.Producer
	delivery chan ckafka.Event
	log      SugaredLogger
}

func NewKafkaPublisher(bootstrapServers, clientID string) (*KafkaPublisher, error) {
	conf := &ckafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         clientID,
		"acks":              "all",
	}
	p, err := ckafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}
	kp := &KafkaPublisher{p: p, delivery: make(chan ckafka.Event, 100)}
	go func() {
		for evt := range kp.delivery {
			if m, ok := evt.(*ckafka.Message); ok {
				if m.TopicPartition.Error != nil {
					if kp.log != nil {
						t := ""
						if m.TopicPartition.Topic != nil {
							t = *m.TopicPartition.Topic
						}
						kp.log.Errorw("kafka delivery failed", "topic", t, "partition", m.TopicPartition.Partition, "offset", m.TopicPartition.Offset, "error", m.TopicPartition.Error)
					}
				} else {
					if kp.log != nil {
						t := ""
						if m.TopicPartition.Topic != nil {
							t = *m.TopicPartition.Topic
						}
						kp.log.Infow("kafka delivery succeeded", "topic", t, "partition", m.TopicPartition.Partition, "offset", m.TopicPartition.Offset)
					}
				}
			}
		}
	}()
	return kp, nil
}

func (k *KafkaPublisher) WithLogger(l SugaredLogger) Publisher { k.log = l; return k }

func (k *KafkaPublisher) Close() error {
	// Ensure all outstanding messages are delivered before closing resources
	_ = k.p.Flush(5000)
	// Signal delivery goroutine to stop after buffered events are drained
	close(k.delivery)
	// Now close the producer
	k.p.Close()
	return nil
}

// Publish is non-blocking: it enqueues the message for delivery and returns immediately.
// Delivery success/failure is logged by the background goroutine.
func (k *KafkaPublisher) Publish(ctx context.Context, topic string, value []byte) error {
	msg := &ckafka.Message{TopicPartition: ckafka.TopicPartition{Topic: &topic, Partition: ckafka.PartitionAny}, Value: value}
	if err := k.p.Produce(msg, k.delivery); err != nil {
		return err
	}
	return nil
}
