package consumer

import (
	"context"

	kafkaclient "github.com/kubernetestest/ecommerce-platform/pkg/kafkaclient"
	"github.com/kubernetestest/ecommerce-platform/pkg/logger"

	events "github.com/kubernetestest/ecommerce-platform/proto-go/events"

	"google.golang.org/protobuf/proto"
)

type OrderCreatedHandler interface {
	Handle(ctx context.Context, evt *events.OrderCreated) error
}

type OrderCreatedHandlerFunc func(ctx context.Context, evt *events.OrderCreated) error

func (f OrderCreatedHandlerFunc) Handle(ctx context.Context, evt *events.OrderCreated) error {
	return f(ctx, evt)
}

type OrderCreatedConsumer struct {
	c   *kafkaclient.Consumer
	h   OrderCreatedHandler
	log logger.Logger
}

func NewOrderCreatedConsumer(bootstrapServers, groupID string, handler OrderCreatedHandler) (*OrderCreatedConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  "earliest",
	}

	c, err := kafkaclient.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	return &OrderCreatedConsumer{c: c, h: handler}, nil
}

func (c *OrderCreatedConsumer) WithLogger(l logger.Logger) *OrderCreatedConsumer {
	c.log = l
	c.c.WithLogger(l)
	return c
}

func (c *OrderCreatedConsumer) Close() error { return c.c.Close() }

func (c *OrderCreatedConsumer) Run(ctx context.Context, topics []string) error {
	return c.c.RunValueLoop(ctx, topics, func(hctx context.Context, value []byte) error {
		var evt events.OrderCreated
		if err := proto.Unmarshal(value, &evt); err != nil {
			return err
		}
		return c.h.Handle(hctx, &evt)
	})
}
