package productinfo

import (
	"context"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// ProductInfo is a simple DTO used by order-service.
type ProductInfo struct {
	Name  string
	Price *valueobjects.Money
}

// Provider abstracts product information lookup (e.g., via inventory-service).
type Provider interface {
	GetProduct(ctx context.Context, productID string) (*ProductInfo, error)
}
