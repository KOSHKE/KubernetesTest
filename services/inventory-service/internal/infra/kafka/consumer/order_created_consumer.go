package consumer

import (
	"context"

	shared "kubernetetest/pkg/kafka"

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

type Consumer struct {
	c *shared.Consumer
	h OrderCreatedHandler
	l shared.SugaredLogger
}

func NewConsumer(bootstrapServers, groupID string, handler OrderCreatedHandler) (*Consumer, error) {
	c, err := shared.NewConsumer(bootstrapServers, groupID, "earliest")
	if err != nil {
		return nil, err
	}
	return &Consumer{c: c, h: handler}, nil
}

func (c *Consumer) WithLogger(l shared.SugaredLogger) *Consumer { c.l = l; c.c.WithLogger(l); return c }

func (c *Consumer) Close() error { return c.c.Close() }

func (c *Consumer) Run(ctx context.Context, topics []string) error {
	return c.c.RunValueLoop(ctx, topics, func(hctx context.Context, value []byte) error {
		var evt events.OrderCreated
		if err := proto.Unmarshal(value, &evt); err != nil {
			return err
		}
		return c.h.Handle(hctx, &evt)
	})
}
