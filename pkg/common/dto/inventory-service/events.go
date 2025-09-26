package dto

// StockEventDTO represents a stock event for outbox
type StockEventDTO struct {
	OrderID string           `json:"order_id"`
	Items   []StockEventItem `json:"items"`
}

// StockEventItem represents an item in stock event
type StockEventItem struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}
