package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ecommerce-platform/pkg/outbox"
)

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
	var records []struct {
		ID          uint
		AggregateID string
		Type        string
		Payload     string
		RetryCount  int
		Processed   bool
		ProcessedAt *time.Time
		FailedAt    *time.Time
		Error       string
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	err := db.Table("outbox_events").Find(&records).Error
	require.NoError(t, err)

	events := make([]outbox.Event, len(records))
	for i, record := range records {
		events[i] = outbox.Event{
			ID:          record.ID,
			AggregateID: record.AggregateID,
			Type:        record.Type,
			Payload:     record.Payload,
			RetryCount:  record.RetryCount,
			ProcessedAt: record.ProcessedAt,
			FailedAt:    record.FailedAt,
			Error:       record.Error,
			CreatedAt:   record.CreatedAt,
		}
	}

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

// AssertPaymentEventPayload checks the payment event payload structure
func AssertPaymentEventPayload(t *testing.T, event outbox.Event, expectedOrderID, expectedPaymentID, expectedUserID string, expectedSuccess bool) {
	// Parse payload to check structure
	payload, ok := event.Payload.(map[string]interface{})
	require.True(t, ok, "Payload should be a map")

	// Check required fields
	assert.Equal(t, expectedOrderID, payload["OrderID"], "OrderID should match")
	assert.Equal(t, expectedPaymentID, payload["PaymentID"], "PaymentID should match")
	assert.Equal(t, expectedUserID, payload["UserID"], "UserID should match")
	assert.Equal(t, expectedSuccess, payload["Success"], "Success should match")
	assert.NotEmpty(t, payload["Message"], "Message should not be empty")
}
