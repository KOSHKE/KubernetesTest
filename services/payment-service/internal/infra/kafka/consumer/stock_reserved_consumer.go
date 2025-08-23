package consumer

import (
	"context"

	kafkaclient "github.com/kubernetestest/ecommerce-platform/pkg/kafkaclient"
	"github.com/kubernetestest/ecommerce-platform/pkg/logger"

	events "github.com/kubernetestest/ecommerce-platform/proto-go/events"

	"google.golang.org/protobuf/proto"
)

type StockReservedHandler interface {
	Handle(ctx context.Context, evt *events.StockReserved) error
}

type StockReservedHandlerFunc func(ctx context.Context, evt *events.StockReserved) error

func (f StockReservedHandlerFunc) Handle(ctx context.Context, evt *events.StockReserved) error {
	return f(ctx, evt)
}

type Consumer struct {
	c   *kafkaclient.Consumer
	h   StockReservedHandler
	log logger.Logger
}

func NewConsumer(bootstrapServers, groupID string, handler StockReservedHandler) (*Consumer, error) {
	config := kafkaclient.ConsumerConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  "earliest",
	}

	c, err := kafkaclient.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	return &Consumer{c: c, h: handler}, nil
}

func (c *Consumer) WithLogger(l logger.Logger) *Consumer {
	c.log = l
	c.c.WithLogger(l)
	return c
}

func (c *Consumer) Close() error { return c.c.Close() }

func (c *Consumer) Run(ctx context.Context, topics []string) error {
	return c.c.RunValueLoop(ctx, topics, func(hctx context.Context, value []byte) error {
		var evt events.StockReserved
		if err := proto.Unmarshal(value, &evt); err != nil {
			return err
		}
		return c.h.Handle(hctx, &evt)
	})
}
