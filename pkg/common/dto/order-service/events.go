package dto

import "ecommerce-platform/proto-go/common"

// OrderEventDTO represents order event data for outbox pattern
type OrderEventDTO struct {
	UserID      string              `json:"user_id"`
	Items       []*common.OrderItem `json:"items"`
	TotalAmount int64               `json:"total_amount"`
	Currency    string              `json:"currency"`
	Reason      string              `json:"reason,omitempty"` // for OrderCancelled events
}
