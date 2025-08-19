package publisher

import (
	"context"

	shared "kubernetetest/libs/kafka"

	"google.golang.org/protobuf/proto"

	events "proto-go/events"
)

type SugaredLogger = shared.SugaredLogger

type Publisher = shared.Publisher

type StockEventsPublisher struct {
	base          Publisher
	topicReserved string
	topicFailed   string
}

func NewStockEventsPublisher(bootstrapServers, topicReserved, topicFailed string) (*StockEventsPublisher, error) {
	kp, err := shared.NewKafkaPublisher(bootstrapServers, "inventory-service")
	if err != nil {
		return nil, err
	}
	return &StockEventsPublisher{base: kp, topicReserved: topicReserved, topicFailed: topicFailed}, nil
}

func (p *StockEventsPublisher) WithLogger(l SugaredLogger) *StockEventsPublisher {
	p.base = p.base.WithLogger(l)
	return p
}

func (p *StockEventsPublisher) Close() error { return p.base.Close() }

func (p *StockEventsPublisher) PublishStockReserved(ctx context.Context, evt *events.StockReserved) error {
	bytes, err := proto.Marshal(evt)
	if err != nil {
		return err
	}
	return p.base.Publish(ctx, p.topicReserved, bytes)
}

func (p *StockEventsPublisher) PublishStockReservationFailed(ctx context.Context, evt *events.StockReservationFailed) error {
	bytes, err := proto.Marshal(evt)
	if err != nil {
		return err
	}
	return p.base.Publish(ctx, p.topicFailed, bytes)
}
