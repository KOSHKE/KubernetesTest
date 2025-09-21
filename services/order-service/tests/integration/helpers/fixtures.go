package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/application/dto"
	"ecommerce-platform/services/order-service/internal/application/services"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// CreateTestOrder creates a test order with the given parameters
func CreateTestOrder(t *testing.T, ctx context.Context, appService *services.OrderApplicationService, userID, currency string, items []dto.OrderItemRequest) string {
	shippingAddr, err := orderValueObjects.NewShippingAddress("123 Test Street, Test City, Test Country")
	require.NoError(t, err)

	curr, err := valueobjects.NewCurrency(currency)
	require.NoError(t, err)

	req := &dto.CreateOrderRequest{
		UserID:          userID,
		Items:           items,
		ShippingAddress: shippingAddr,
		Currency:        curr,
	}

	response, err := appService.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, response.ID)

	return response.ID
}

// GetOrderByID retrieves an order by ID for testing
func GetOrderByID(t *testing.T, ctx context.Context, appService *services.OrderApplicationService, orderID, userID string) *dto.OrderResponse {
	req := &dto.GetOrderRequest{
		OrderID: orderID,
		UserID:  userID,
	}

	response, err := appService.GetOrder(ctx, req)
	require.NoError(t, err)
	return response
}

// MustCurrency creates a currency or panics (for test fixtures)
func MustCurrency(code string) valueobjects.Currency {
	curr, err := valueobjects.NewCurrency(code)
	if err != nil {
		panic(err)
	}
	return curr
}

// CreateOrderItemRequest creates a test order item request
func CreateOrderItemRequest(productID, productName string, quantity int32, priceAmount int64, currency string) dto.OrderItemRequest {
	return dto.OrderItemRequest{
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		Price:       valueobjects.NewMoney(priceAmount, MustCurrency(currency)),
	}
}
