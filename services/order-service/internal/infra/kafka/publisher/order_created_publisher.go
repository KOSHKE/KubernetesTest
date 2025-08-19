package publisher

import (
	"context"

	shared "kubernetetest/libs/kafka"

	"google.golang.org/protobuf/proto"

	"proto-go/events"
)

type SugaredLogger = shared.SugaredLogger

type Publisher = shared.Publisher

type OrderCreatedPublisher struct {
	base  Publisher
	topic string
}

func NewOrderCreatedPublisher(bootstrapServers, topic string) (*OrderCreatedPublisher, error) {
	kp, err := shared.NewKafkaPublisher(bootstrapServers, "order-service")
	if err != nil {
		return nil, err
	}
	return &OrderCreatedPublisher{base: kp, topic: topic}, nil
}

func (p *OrderCreatedPublisher) WithLogger(l SugaredLogger) *OrderCreatedPublisher {
	p.base = p.base.WithLogger(l)
	return p
}

func (p *OrderCreatedPublisher) Close() error { return p.base.Close() }

func (p *OrderCreatedPublisher) PublishOrderCreated(ctx context.Context, evt *events.OrderCreated) error {
	bytes, err := proto.Marshal(evt)
	if err != nil {
		return err
	}
	return p.base.Publish(ctx, p.topic, bytes)
}
