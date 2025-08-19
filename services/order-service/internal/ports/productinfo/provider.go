package productinfo

import "context"

// ProductInfo is a simple DTO used by order-service.
type ProductInfo struct {
	Name     string
	Price    int64  // minor units
	Currency string // ISO code
}

// Provider abstracts product information lookup (e.g., via inventory-service).
type Provider interface {
	GetProduct(ctx context.Context, productID string) (*ProductInfo, error)
}
