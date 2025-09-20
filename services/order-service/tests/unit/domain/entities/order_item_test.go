package entities_test

import (
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestOrderItem_NewOrderItem(t *testing.T) {
	t.Run("valid order item creation", func(t *testing.T) {
		t.Parallel()
		// Arrange
		productID := "product-123"
		productName := "Test Product"
		quantity := int32(2)
		unitPrice := createTestMoney(1000, "USD")

		// Act
		item := entities.NewOrderItem(productID, productName, quantity, unitPrice)

		// Assert
		assert.Equal(t, productID, item.ProductID)
		assert.Equal(t, productName, item.ProductName)
		assert.Equal(t, quantity, item.Quantity)
		assert.Equal(t, unitPrice, item.UnitPrice)
	})

	t.Run("order item with zero quantity", func(t *testing.T) {
		t.Parallel()
		// Arrange
		productID := "product-456"
		productName := "Zero Quantity Product"
		quantity := int32(0)
		unitPrice := createTestMoney(500, "EUR")

		// Act
		item := entities.NewOrderItem(productID, productName, quantity, unitPrice)

		// Assert
		assert.Equal(t, productID, item.ProductID)
		assert.Equal(t, productName, item.ProductName)
		assert.Equal(t, quantity, item.Quantity)
		assert.Equal(t, unitPrice, item.UnitPrice)
	})
}

func TestOrderItem_ChangeQuantity(t *testing.T) {
	t.Run("change quantity to positive value", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-123", "Test Product", 2, 1000, "USD")
		newQuantity := int32(5)

		// Act
		item.ChangeQuantity(newQuantity)

		// Assert
		assert.Equal(t, newQuantity, item.Quantity)
	})

	t.Run("change quantity to zero", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-123", "Test Product", 3, 1000, "USD")
		newQuantity := int32(0)

		// Act
		item.ChangeQuantity(newQuantity)

		// Assert
		assert.Equal(t, newQuantity, item.Quantity)
	})
}

func TestOrderItem_TotalPrice(t *testing.T) {
	t.Run("calculate total price for multiple items", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-123", "Test Product", 3, 1000, "USD")
		expectedTotal := createTestMoney(3000, "USD")

		// Act
		totalPrice := item.TotalPrice()

		// Assert
		assert.Equal(t, expectedTotal, totalPrice)
	})

	t.Run("calculate total price for single item", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-456", "Single Item", 1, 2500, "EUR")
		expectedTotal := createTestMoney(2500, "EUR")

		// Act
		totalPrice := item.TotalPrice()

		// Assert
		assert.Equal(t, expectedTotal, totalPrice)
	})

	t.Run("calculate total price for zero quantity", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-789", "Zero Quantity", 0, 1500, "USD")
		expectedTotal := createTestMoney(0, "USD")

		// Act
		totalPrice := item.TotalPrice()

		// Assert
		assert.Equal(t, expectedTotal, totalPrice)
	})
}

func TestOrderItem_Currency(t *testing.T) {
	t.Run("get currency code", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-123", "Test Product", 2, 1000, "EUR")

		// Act
		currency := item.Currency()

		// Assert
		assert.Equal(t, "EUR", currency)
	})
}

func TestOrderItem_UnitPriceAmount(t *testing.T) {
	t.Run("get unit price amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-123", "Test Product", 2, 1500, "USD")

		// Act
		amount := item.UnitPriceAmount()

		// Assert
		assert.Equal(t, int64(1500), amount)
	})
}

func TestOrderItem_TotalPriceAmount(t *testing.T) {
	t.Run("get total price amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-123", "Test Product", 4, 750, "USD")

		// Act
		totalAmount := item.TotalPriceAmount()

		// Assert
		assert.Equal(t, int64(3000), totalAmount) // 4 * 750 = 3000
	})

	t.Run("get total price amount for zero quantity", func(t *testing.T) {
		t.Parallel()
		// Arrange
		item := createTestOrderItem("product-456", "Zero Item", 0, 1000, "EUR")

		// Act
		totalAmount := item.TotalPriceAmount()

		// Assert
		assert.Equal(t, int64(0), totalAmount)
	})
}

// Helper functions for creating test data
func createTestOrderItem(productID, productName string, quantity int32, amount int64, currency string) *entities.OrderItem {
	unitPrice := createTestMoney(amount, currency)
	return entities.NewOrderItem(productID, productName, quantity, unitPrice)
}

func createTestMoney(amount int64, currency string) valueobjects.Money {
	currencyObj, _ := valueobjects.NewCurrency(currency)
	return valueobjects.NewMoney(amount, currencyObj)
}
