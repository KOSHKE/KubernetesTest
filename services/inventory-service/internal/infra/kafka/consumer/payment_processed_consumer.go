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

type PaymentConsumer struct {
	c *shared.Consumer
	h PaymentProcessedHandler
	l shared.SugaredLogger
}

func NewPaymentConsumer(bootstrapServers, groupID string, handler PaymentProcessedHandler) (*PaymentConsumer, error) {
	c, err := shared.NewConsumer(bootstrapServers, groupID, "earliest")
	if err != nil {
		return nil, err
	}
	return &PaymentConsumer{c: c, h: handler}, nil
}

func (c *PaymentConsumer) WithLogger(l shared.SugaredLogger) *PaymentConsumer {
	c.l = l
	c.c.WithLogger(l)
	return c
}

func (c *PaymentConsumer) Close() error { return c.c.Close() }

func (c *PaymentConsumer) Run(ctx context.Context, topics []string) error {
	return c.c.RunValueLoop(ctx, topics, func(hctx context.Context, value []byte) error {
		var evt events.PaymentProcessed
		if err := proto.Unmarshal(value, &evt); err != nil {
			return err
		}
		return c.h.Handle(hctx, &evt)
	})
}
