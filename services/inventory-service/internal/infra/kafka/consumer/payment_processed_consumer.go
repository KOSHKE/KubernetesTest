package consumer

import (
	"context"

	kafkaclient "github.com/kubernetestest/ecommerce-platform/pkg/kafkaclient"
	"github.com/kubernetestest/ecommerce-platform/pkg/logger"

	events "github.com/kubernetestest/ecommerce-platform/proto-go/events"

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
	c *kafkaclient.Consumer
	h PaymentProcessedHandler
	l logger.Logger
}

func NewPaymentConsumer(bootstrapServers, groupID string, handler PaymentProcessedHandler) (*PaymentConsumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  "earliest",
	}

	c, err := kafkaclient.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	return &PaymentConsumer{c: c, h: handler}, nil
}

func (c *PaymentConsumer) WithLogger(l logger.Logger) *PaymentConsumer {
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
