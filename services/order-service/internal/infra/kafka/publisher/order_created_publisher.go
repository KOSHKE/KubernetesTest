package publisher

import (
	"context"

	kafkaclient "github.com/kubernetestest/ecommerce-platform/pkg/kafkaclient"
	"github.com/kubernetestest/ecommerce-platform/pkg/logger"
	"google.golang.org/protobuf/proto"

	"github.com/kubernetestest/ecommerce-platform/proto-go/events"
)

type Publisher = kafkaclient.Publisher

type OrderCreatedPublisher struct {
	base  Publisher
	topic string
}

func NewOrderCreatedPublisher(bootstrapServers, topic string) (*OrderCreatedPublisher, error) {
	config := kafkaclient.PublisherConfig{
		BootstrapServers: bootstrapServers,
		ClientID:         "order-service",
	}

	kp, err := kafkaclient.NewKafkaPublisher(config)
	if err != nil {
		return nil, err
	}
	return &OrderCreatedPublisher{base: kp, topic: topic}, nil
}

func (p *OrderCreatedPublisher) WithLogger(l logger.Logger) *OrderCreatedPublisher {
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
