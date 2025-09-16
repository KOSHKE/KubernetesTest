package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/application/services"
)

// AssertStockReserved checks that stock is reserved
func AssertStockReserved(t *testing.T, ctx context.Context, appService *services.InventoryApplicationService, productID string, expectedAvailable, expectedReserved int32) {
	stock, err := appService.GetStockByProductID(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, expectedAvailable, stock.AvailableQuantity, "Available quantity should match")
	assert.Equal(t, expectedReserved, stock.ReservedQuantity, "Reserved quantity should match")
}

// AssertOutboxEventExists checks that event exists in outbox
func AssertOutboxEventExists(t *testing.T, ctx context.Context, db *gorm.DB, eventType, aggregateID string) {
	events := GetOutboxEvents(t, ctx, db)

	var found bool
	for _, event := range events {
		if event.Type == eventType && event.AggregateID == aggregateID {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected outbox event not found: type=%s, aggregateID=%s", eventType, aggregateID)
}

// AssertOutboxEventCount checks the number of events in outbox
func AssertOutboxEventCount(t *testing.T, ctx context.Context, db *gorm.DB, expectedCount int) {
	events := GetOutboxEvents(t, ctx, db)
	assert.Len(t, events, expectedCount, "Outbox event count should match")
}

// AssertOutboxEventsByType checks events of specific type
func AssertOutboxEventsByType(t *testing.T, ctx context.Context, db *gorm.DB, eventType string, expectedCount int) {
	events := GetOutboxEvents(t, ctx, db)
	filtered := FilterEventsByType(events, eventType)
	assert.Len(t, filtered, expectedCount, "Expected %d events of type %s", expectedCount, eventType)
}

// GetOutboxEvents gets all events from outbox
func GetOutboxEvents(t *testing.T, ctx context.Context, db *gorm.DB) []outbox.Event {
	var events []outbox.Event
	err := db.Find(&events).Error
	require.NoError(t, err)
	return events
}

// FilterEventsByType filters events by type
func FilterEventsByType(events []outbox.Event, eventType string) []outbox.Event {
	var filtered []outbox.Event
	for _, event := range events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// AssertEventProcessed checks that event is marked as processed
func AssertEventProcessed(t *testing.T, ctx context.Context, db *gorm.DB, eventType, aggregateID string) {
	events := GetOutboxEvents(t, ctx, db)

	var found bool
	for _, event := range events {
		if event.Type == eventType && event.AggregateID == aggregateID {
			assert.True(t, event.ProcessedAt != nil, "Event should be marked as processed")
			found = true
			break
		}
	}

	assert.True(t, found, "Event not found: type=%s, aggregateID=%s", eventType, aggregateID)
}
