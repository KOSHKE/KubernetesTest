package kafka

import (
	"context"

	events "proto-go/events"

	"google.golang.org/protobuf/proto"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	p     *ckafka.Producer
	topic string
}

func NewProducer(bootstrapServers, topic string) (*Producer, error) {
	conf := &ckafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"client.id":         "order-service",
		"acks":              "all",
	}
	p, err := ckafka.NewProducer(conf)
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
	msg := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &p.topic, Partition: ckafka.PartitionAny},
		Value:          bytes,
	}

	// support external timeout/cancel via ctx
	done := make(chan error, 1)
	go func() {
		delivery := make(chan ckafka.Event, 1)
		defer close(delivery)
		if err := p.p.Produce(msg, delivery); err != nil {
			done <- err
			return
		}
		e := <-delivery
		m := e.(*ckafka.Message)
		done <- m.TopicPartition.Error
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
