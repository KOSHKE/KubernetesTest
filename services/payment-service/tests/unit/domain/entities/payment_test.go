package entities_test

import (
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/payment-service/internal/domain/entities"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"

	"github.com/stretchr/testify/assert"
)

func TestPayment_NewPayment(t *testing.T) {
	t.Run("valid payment creation", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orderID := "order-123"
		userID := "user-456"
		amount := createTestMoney(1000, "USD")
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment := entities.NewPayment(orderID, userID, amount, method)

		// Assert
		assert.NotEmpty(t, payment.ID)
		assert.Equal(t, orderID, payment.OrderID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, amount, payment.Amount)
		assert.Equal(t, method, payment.Method)
		assert.Equal(t, paymentvalueobjects.PaymentStatusPending, payment.Status)
		assert.Empty(t, payment.TransactionID)
		assert.False(t, payment.CreatedAt.IsZero())
		assert.False(t, payment.UpdatedAt.IsZero())
	})

	t.Run("payment with negative amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orderID := "order-999"
		userID := "user-888"
		amount := createTestMoney(-100, "USD")
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment := entities.NewPayment(orderID, userID, amount, method)

		// Assert
		assert.NotEmpty(t, payment.ID)
		assert.Equal(t, orderID, payment.OrderID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, amount, payment.Amount)
		assert.Equal(t, method, payment.Method)
		assert.Equal(t, paymentvalueobjects.PaymentStatusPending, payment.Status)
	})

	t.Run("payment with empty order ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orderID := ""
		userID := "user-123"
		amount := createTestMoney(1000, "USD")
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment := entities.NewPayment(orderID, userID, amount, method)

		// Assert
		assert.NotEmpty(t, payment.ID)
		assert.Equal(t, orderID, payment.OrderID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, amount, payment.Amount)
		assert.Equal(t, method, payment.Method)
		assert.Equal(t, paymentvalueobjects.PaymentStatusPending, payment.Status)
	})

	t.Run("payment with empty user ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		orderID := "order-123"
		userID := ""
		amount := createTestMoney(1000, "USD")
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment := entities.NewPayment(orderID, userID, amount, method)

		// Assert
		assert.NotEmpty(t, payment.ID)
		assert.Equal(t, orderID, payment.OrderID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, amount, payment.Amount)
		assert.Equal(t, method, payment.Method)
		assert.Equal(t, paymentvalueobjects.PaymentStatusPending, payment.Status)
	})
}

func TestPayment_CanBeProcessed(t *testing.T) {
	t.Run("pending payment can be processed", func(t *testing.T) {
		t.Parallel()
		// Arrange
		payment := createTestPayment("order-123", "user-456", 1000, "USD")

		// Act
		err := payment.CanBeProcessed()

		// Assert
		assert.NoError(t, err)
	})
}

func TestPayment_Validation(t *testing.T) {
	t.Run("valid payment", func(t *testing.T) {
		t.Parallel()
		// Arrange
		payment := createTestPayment("order-123", "user-456", 1000, "USD")

		// Act
		err := payment.CanBeProcessed()

		// Assert
		assert.NoError(t, err)
	})
}

// Helper functions for creating test data
func createTestPayment(orderID, userID string, amount int64, currency string) *entities.Payment {
	currencyObj, _ := valueobjects.NewCurrency(currency)
	money := valueobjects.NewMoney(amount, currencyObj)
	return entities.NewPayment(orderID, userID, money, paymentvalueobjects.PaymentMethodCreditCard)
}

func createTestMoney(amount int64, currency string) valueobjects.Money {
	currencyObj, _ := valueobjects.NewCurrency(currency)
	return valueobjects.NewMoney(amount, currencyObj)
}
