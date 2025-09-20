package aggregates_test

import (
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"

	"github.com/stretchr/testify/assert"
)

func TestOrder_NewOrder(t *testing.T) {
	t.Run("valid order creation", func(t *testing.T) {
		t.Parallel()
		// Arrange
		userID := "user-123"
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
		currency, _ := valueobjects.NewCurrency("USD")

		// Act
		order, err := aggregates.NewOrder(userID, shippingAddress, currency)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, userID, order.UserID)
		assert.Equal(t, shippingAddress, order.ShippingAddress)
		assert.Equal(t, currency, order.Currency)
		assert.Equal(t, orderValueObjects.OrderStatusPending, order.Status)
		assert.Empty(t, order.Items)
		assert.Equal(t, valueobjects.NewMoney(0, currency), order.TotalAmount)
	})
}

func TestOrder_AddItem(t *testing.T) {
	t.Run("add new item to empty order", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		productID := "product-456"
		productName := "Test Product"
		quantity := int32(2)
		unitPrice := createTestMoney(1000, "USD")

		// Act
		err := order.AddItem(productID, productName, quantity, unitPrice)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, order.Items, 1)
		assert.Equal(t, productID, order.Items[0].ProductID)
		assert.Equal(t, productName, order.Items[0].ProductName)
		assert.Equal(t, quantity, order.Items[0].Quantity)
		assert.Equal(t, unitPrice, order.Items[0].UnitPrice)
		assert.Equal(t, createTestMoney(2000, "USD"), order.TotalAmount)
	})

	t.Run("add same product twice - should update quantity", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		productID := "product-456"
		productName := "Test Product"
		quantity1 := int32(2)
		quantity2 := int32(3)
		unitPrice := createTestMoney(1000, "USD")

		// Act
		err1 := order.AddItem(productID, productName, quantity1, unitPrice)
		err2 := order.AddItem(productID, productName, quantity2, unitPrice)

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Len(t, order.Items, 1)                                    // Should still be one item
		assert.Equal(t, quantity1+quantity2, order.Items[0].Quantity)    // Quantity should be combined
		assert.Equal(t, createTestMoney(5000, "USD"), order.TotalAmount) // 5 * 1000
	})

	t.Run("add item with different currency - should fail", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		productID := "product-456"
		productName := "Test Product"
		quantity := int32(1)
		unitPrice := createTestMoney(1000, "EUR") // Different currency

		// Act
		err := order.AddItem(productID, productName, quantity, unitPrice)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrCurrencyMismatch, err)
		assert.Empty(t, order.Items)
	})
}

func TestOrder_RemoveItem(t *testing.T) {
	t.Run("remove existing item", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		productID := "product-456"
		_ = order.AddItem(productID, "Test Product", 2, createTestMoney(1000, "USD"))

		// Act
		err := order.RemoveItem(productID)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, order.Items)
		assert.Equal(t, createTestMoney(0, "USD"), order.TotalAmount)
	})

	t.Run("remove non-existent item", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		_ = order.AddItem("product-456", "Test Product", 2, createTestMoney(1000, "USD"))

		// Act
		err := order.RemoveItem("non-existent-product")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderItemNotFound, err)
		assert.Len(t, order.Items, 1) // Original item should remain
	})
}

func TestOrder_SetStatus(t *testing.T) {
	t.Run("set status to confirmed", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		newStatus := orderValueObjects.OrderStatusConfirmed

		// Act
		err := order.SetStatus(newStatus)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, newStatus, order.Status)
	})

	t.Run("set status to cancelled", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		newStatus := orderValueObjects.OrderStatusCancelled

		// Act
		err := order.SetStatus(newStatus)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, newStatus, order.Status)
	})
}

func TestOrder_CancelOrder(t *testing.T) {
	t.Run("cancel order by owner", func(t *testing.T) {
		t.Parallel()
		// Arrange
		userID := "user-123"
		order := createTestOrder(userID, "USD")

		// Act
		err := order.CancelOrder(userID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, orderValueObjects.OrderStatusCancelled, order.Status)
	})

	t.Run("cancel order by different user - should fail", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		differentUserID := "user-456"

		// Act
		err := order.CancelOrder(differentUserID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderAccessDenied, err)
		assert.Equal(t, orderValueObjects.OrderStatusPending, order.Status) // Status should remain unchanged
	})

	t.Run("cancel already confirmed order - should fail", func(t *testing.T) {
		t.Parallel()
		// Arrange
		userID := "user-123"
		order := createTestOrder(userID, "USD")
		_ = order.SetStatus(orderValueObjects.OrderStatusConfirmed)

		// Act
		err := order.CancelOrder(userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderCancellationFailed, err)
		assert.Equal(t, orderValueObjects.OrderStatusConfirmed, order.Status) // Status should remain confirmed
	})
}

func TestOrder_IsOwnedBy(t *testing.T) {
	t.Run("check ownership by correct user", func(t *testing.T) {
		t.Parallel()
		// Arrange
		userID := "user-123"
		order := createTestOrder(userID, "USD")

		// Act
		isOwned := order.IsOwnedBy(userID)

		// Assert
		assert.True(t, isOwned)
	})

	t.Run("check ownership by different user", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")
		differentUserID := "user-456"

		// Act
		isOwned := order.IsOwnedBy(differentUserID)

		// Assert
		assert.False(t, isOwned)
	})
}

func TestOrder_RecalculateTotal(t *testing.T) {
	t.Run("total calculation with multiple items", func(t *testing.T) {
		t.Parallel()
		// Arrange
		order := createTestOrder("user-123", "USD")

		// Add multiple items
		_ = order.AddItem("product-1", "Product 1", 2, createTestMoney(1000, "USD")) // 2000
		_ = order.AddItem("product-2", "Product 2", 1, createTestMoney(1500, "USD")) // 1500

		// Total should be 3500
		expectedTotal := createTestMoney(3500, "USD")

		// Assert
		assert.Equal(t, expectedTotal, order.TotalAmount)
		assert.Len(t, order.Items, 2)
	})
}

// Helper functions for creating test data
func createTestOrder(userID, currency string) *aggregates.Order {
	shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
	currencyObj, _ := valueobjects.NewCurrency(currency)
	order, _ := aggregates.NewOrder(userID, shippingAddress, currencyObj)
	return order
}

func createTestMoney(amount int64, currency string) valueobjects.Money {
	currencyObj, _ := valueobjects.NewCurrency(currency)
	return valueobjects.NewMoney(amount, currencyObj)
}
