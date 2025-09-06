package publisher

import (
	"context"

	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/publisher"
)

// StockEventsPublisherImpl implements the StockEventsPublisher interface
type StockEventsPublisherImpl struct {
	reservedPublisher  *StockReservedPublisher
	releasedPublisher  *StockReleasedPublisher
	committedPublisher *StockCommittedPublisher
}

// NewStockEventsPublisher creates a new stock events publisher
func NewStockEventsPublisher(brokers []string, reservedTopic, releasedTopic, committedTopic string) (publisher.StockEventsPublisher, error) {
	// Create individual publishers
	reservedPublisher, err := NewStockReservedPublisher(brokers[0], reservedTopic)
	if err != nil {
		return nil, err
	}

	releasedPublisher, err := NewStockReleasedPublisher(brokers[0], releasedTopic)
	if err != nil {
		return nil, err
	}

	committedPublisher, err := NewStockCommittedPublisher(brokers[0], committedTopic)
	if err != nil {
		return nil, err
	}

	return &StockEventsPublisherImpl{
		reservedPublisher:  reservedPublisher,
		releasedPublisher:  releasedPublisher,
		committedPublisher: committedPublisher,
	}, nil
}

// PublishStockReserved publishes when stock is successfully reserved
func (p *StockEventsPublisherImpl) PublishStockReserved(ctx context.Context, event *events.StockReserved) error {
	return p.reservedPublisher.PublishStockReserved(ctx, event)
}

// PublishStockReleased publishes when reserved stock is released
func (p *StockEventsPublisherImpl) PublishStockReleased(ctx context.Context, event *events.StockReleased) error {
	return p.releasedPublisher.PublishStockReleased(ctx, event)
}

// PublishStockCommitted publishes when reserved stock is committed
func (p *StockEventsPublisherImpl) PublishStockCommitted(ctx context.Context, event *events.StockCommitted) error {
	return p.committedPublisher.PublishStockCommitted(ctx, event)
}

// Close closes the publisher
func (p *StockEventsPublisherImpl) Close() error {
	var firstErr error

	if err := p.reservedPublisher.Close(); err != nil {
		firstErr = err
	}

	if err := p.releasedPublisher.Close(); err != nil {
		if firstErr == nil {
			firstErr = err
		}
	}

	if err := p.committedPublisher.Close(); err != nil {
		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
