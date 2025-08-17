package kafka

import (
	"context"
	"time"

	events "proto-go/events"

	"google.golang.org/protobuf/proto"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	p     *kafka.Producer
	topic string
}

func NewProducer(bootstrapServers, topic string) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "order-service",
		"acks":              "all",
	}
	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}
	return &Producer{p: p, topic: topic}, nil
}

func (p *Producer) Close() { p.p.Close() }

func (p *Producer) PublishOrderCreated(ctx context.Context, evt *events.OrderCreated) error {
	// marshal protobuf
	bytes, err := proto.Marshal(evt)
	if err != nil {
		return err
	}
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          bytes,
	}
	delivery := make(chan kafka.Event, 1)
	defer close(delivery)
	if err := p.p.Produce(msg, delivery); err != nil {
		return err
	}
	select {
	case e := <-delivery:
		m := e.(*kafka.Message)
		return m.TopicPartition.Error
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
}
