package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	dto "ecommerce-platform/pkg/common/dto/inventory-service"
	"ecommerce-platform/services/inventory-service/internal/application/services"
)

// CreateTestProduct creates a test product
func CreateTestProduct(t *testing.T, ctx context.Context, appService *services.InventoryApplicationService) *dto.ProductResponse {
	req := &dto.CreateProductRequest{
		Name:        "Test Product",
		PriceAmount: 1000,
		Currency:    "USD",
		ImageURL:    "https://example.com/image.jpg",
		Stock:       nil, // Add stock separately
	}

	product, err := appService.CreateProduct(ctx, req)
	require.NoError(t, err)
	return product
}

// AddTestStock adds test stock to a product
func AddTestStock(t *testing.T, ctx context.Context, appService *services.InventoryApplicationService, productID string, quantity int32) {
	_, err := appService.AddStock(ctx, productID, quantity)
	require.NoError(t, err)
}

// ReserveStockForOrder reserves stock for an order
func ReserveStockForOrder(t *testing.T, ctx context.Context, appService *services.InventoryApplicationService, orderID, productID string, quantity int32) {
	req := &dto.ReserveStockRequest{
		OrderID: orderID,
		Items: []dto.StockReservationItem{
			{
				ProductID: productID,
				Quantity:  quantity,
			},
		},
	}

	_, err := appService.ReserveStock(ctx, req)
	require.NoError(t, err)
}
