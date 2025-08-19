package consumer

import (
	"context"

	shared "kubernetetest/libs/kafka"

	events "proto-go/events"

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
	c   *shared.Consumer
	h   OrderCreatedHandler
	log shared.SugaredLogger
}

func NewOrderCreatedConsumer(bootstrapServers, groupID string, handler OrderCreatedHandler) (*OrderCreatedConsumer, error) {
	c, err := shared.NewConsumer(bootstrapServers, groupID, "earliest")
	if err != nil {
		return nil, err
	}
	return &OrderCreatedConsumer{c: c, h: handler}, nil
}

func (c *OrderCreatedConsumer) WithLogger(l shared.SugaredLogger) *OrderCreatedConsumer {
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
