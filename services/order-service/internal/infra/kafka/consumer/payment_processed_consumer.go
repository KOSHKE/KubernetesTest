package consumer

import (
	"context"

	shared "kubernetetest/libs/kafka"

	events "proto-go/events"

	"google.golang.org/protobuf/proto"
)

type PaymentProcessedHandler interface {
	Handle(ctx context.Context, evt *events.PaymentProcessed) error
}

type PaymentProcessedHandlerFunc func(ctx context.Context, evt *events.PaymentProcessed) error

func (f PaymentProcessedHandlerFunc) Handle(ctx context.Context, evt *events.PaymentProcessed) error {
	return f(ctx, evt)
}

type Consumer struct {
	c   *shared.Consumer
	h   PaymentProcessedHandler
	log shared.SugaredLogger
}

func NewConsumer(bootstrapServers, groupID, autoOffsetReset string, handler PaymentProcessedHandler) (*Consumer, error) {
	c, err := shared.NewConsumer(bootstrapServers, groupID, autoOffsetReset)
	if err != nil {
		return nil, err
	}
	return &Consumer{c: c, h: handler}, nil
}

func (c *Consumer) WithLogger(l shared.SugaredLogger) *Consumer {
	c.log = l
	c.c.WithLogger(l)
	return c
}

func (c *Consumer) Close() error { return c.c.Close() }

func (c *Consumer) Run(ctx context.Context, topics []string) error {
	return c.c.RunValueLoop(ctx, topics, func(hctx context.Context, value []byte) error {
		var evt events.PaymentProcessed
		if err := proto.Unmarshal(value, &evt); err != nil {
			return err
		}
		return c.h.Handle(hctx, &evt)
	})
}
