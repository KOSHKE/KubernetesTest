package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ecommerce-platform/pkg/outbox"
)

// GetOutboxEvents retrieves all outbox events for testing
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

// FilterEventsByType filters outbox events by type
func FilterEventsByType(events []outbox.Event, eventType string) []outbox.Event {
	var filtered []outbox.Event
	for _, event := range events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// FilterEventsByAggregateID filters outbox events by aggregate ID
func FilterEventsByAggregateID(events []outbox.Event, aggregateID string) []outbox.Event {
	var filtered []outbox.Event
	for _, event := range events {
		if event.AggregateID == aggregateID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// AssertEventExists checks if an event of the given type exists for the aggregate
func AssertEventExists(t *testing.T, events []outbox.Event, eventType, aggregateID string) {
	for _, event := range events {
		if event.Type == eventType && event.AggregateID == aggregateID {
			return // Found the event
		}
	}
	t.Fatalf("Expected event of type %s for aggregate %s not found", eventType, aggregateID)
}

// AssertEventProcessed checks if an event was processed
func AssertEventProcessed(t *testing.T, events []outbox.Event, eventType, aggregateID string, processed bool) {
	for _, event := range events {
		if event.Type == eventType && event.AggregateID == aggregateID {
			if processed && event.ProcessedAt == nil {
				t.Fatalf("Expected event %s for aggregate %s to be processed", eventType, aggregateID)
			}
			if !processed && event.ProcessedAt != nil {
				t.Fatalf("Expected event %s for aggregate %s to NOT be processed", eventType, aggregateID)
			}
			return
		}
	}
	t.Fatalf("Event of type %s for aggregate %s not found", eventType, aggregateID)
}
