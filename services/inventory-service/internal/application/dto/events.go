package dto

import "ecommerce-platform/pkg/common/valueobjects"

// StockEventDTO represents a stock event for outbox
type StockEventDTO struct {
	OrderID string              `json:"order_id"`
	UserID  string              `json:"user_id"`
	Items   []valueobjects.Item `json:"items"`
}
