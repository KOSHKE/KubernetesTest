package kafka

import (
	"context"
	"time"

	events "proto-go/events"

	"google.golang.org/protobuf/proto"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	p                  *kafka.Producer
	topicReserved      string
	topicReserveFailed string
}

func NewProducer(bootstrapServers, topicReserved, topicReserveFailed string) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "inventory-service",
		"acks":              "all",
	}
	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}
	return &Producer{p: p, topicReserved: topicReserved, topicReserveFailed: topicReserveFailed}, nil
}

func (p *Producer) Close() { p.p.Close() }

func (p *Producer) PublishStockReserved(ctx context.Context, evt *events.StockReserved) error {
	return p.publish(ctx, p.topicReserved, evt)
}

func (p *Producer) PublishStockReservationFailed(ctx context.Context, evt *events.StockReservationFailed) error {
	return p.publish(ctx, p.topicReserveFailed, evt)
}

func (p *Producer) publish(ctx context.Context, topic string, m proto.Message) error {
	bytes, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	msg := &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: bytes}
	delivery := make(chan kafka.Event, 1)
	defer close(delivery)
	if err := p.p.Produce(msg, delivery); err != nil {
		return err
	}
	select {
	case e := <-delivery:
		km := e.(*kafka.Message)
		return km.TopicPartition.Error
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
}
