package productinfo

import "context"

// Provider abstracts product information lookup (e.g., via inventory-service).
// It returns product name, unit price in minor units and ISO currency.
type Provider interface {
	GetProduct(ctx context.Context, productID string) (name string, price int64, currency string, err error)
}
