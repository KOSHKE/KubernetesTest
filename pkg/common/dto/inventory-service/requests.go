package dto

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	Name        string        `json:"name" validate:"required,min=1,max=255"`
	PriceAmount int64         `json:"price_amount" validate:"required"`
	Currency    string        `json:"currency" validate:"required"`
	ImageURL    string        `json:"image_url" validate:"max=500"`
	Stock       *StockRequest `json:"stock" validate:"omitempty"`
}

// StockRequest represents stock information in request
type StockRequest struct {
	AvailableQuantity int32 `json:"available_quantity" validate:"min=0"`
	ReservedQuantity  int32 `json:"reserved_quantity" validate:"min=0"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=255"`
	PriceAmount int64  `json:"price_amount" validate:"omitempty"`
	Currency    string `json:"currency" validate:"omitempty"`
	ImageURL    string `json:"image_url" validate:"omitempty,max=500"`
}

// ReserveStockRequest represents a request to reserve stock
type ReserveStockRequest struct {
	OrderID string                 `json:"order_id" validate:"required,min=1,max=255"`
	Items   []StockReservationItem `json:"items" validate:"required,min=1,dive"`
}

// StockReservationItem represents an item for stock reservation
type StockReservationItem struct {
	ProductID string `json:"product_id" validate:"required,min=1,max=255"`
	Quantity  int32  `json:"quantity" validate:"required,min=1"`
}

// ListProductsRequest represents a request to list products
type ListProductsRequest struct {
	Page   int    `json:"page" validate:"min=1"`
	Limit  int    `json:"limit" validate:"min=1,max=100"`
	Search string `json:"search" validate:"max=255"`
}

// ReleaseStockRequest represents a request to release stock
type ReleaseStockRequest struct {
	OrderID string                 `json:"order_id" validate:"required,min=1,max=255"`
	Items   []StockReservationItem `json:"items" validate:"required,min=1,dive"`
}

// CommitStockRequest represents a request to commit stock
type CommitStockRequest struct {
	OrderID string                 `json:"order_id" validate:"required,min=1,max=255"`
	Items   []StockReservationItem `json:"items" validate:"required,min=1,dive"`
}
